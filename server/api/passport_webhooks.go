package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/helpers"

	"github.com/ninja-syndicate/ws"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type PassportWebhookController struct {
	API *API
}

func PassportWebhookRouter(webhookSecret string, api *API) chi.Router {
	c := &PassportWebhookController{
		API: api,
	}
	r := chi.NewRouter()
	r.Post("/user_update", WithPassportSecret(webhookSecret, WithError(c.UserUpdated)))
	r.Post("/user_enlist_faction", WithPassportSecret(webhookSecret, WithError(c.UserEnlistFaction)))
	r.Post("/user_stat", WithPassportSecret(webhookSecret, WithError(c.UserStatGet)))
	r.Post("/faction_stat", WithPassportSecret(webhookSecret, WithError(c.FactionStatGet)))

	return r
}

type UserUpdateRequest struct {
	User *server.User `json:"user"`
}

// UserUpdated update user detail
func (pc *PassportWebhookController) UserUpdated(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &UserUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// get player
	player, err := boiler.FindPlayer(gamedb.StdConn, req.User.ID.String())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// update user
	player.Username = null.StringFrom(req.User.Username)
	player.FactionID = null.StringFromPtr(nil)
	player.MobileNumber = req.User.MobileNumber
	if !req.User.FactionID.IsNil() {

		player.FactionID = null.StringFrom(req.User.FactionID.String())

		faction, err := boiler.FindFaction(gamedb.StdConn, req.User.FactionID.String())
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "faction not found")
		}
		err = req.User.Faction.SetFromBoilerFaction(faction)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Unable to convert faction, contact support or try again.")
		}
	}

	req.User.Gid = player.Gid

	// update player
	_, err = player.Update(gamedb.StdConn, boil.Whitelist(
		boiler.PlayerColumns.Username,
		boiler.PlayerColumns.FactionID,
		boiler.PlayerColumns.MobileNumber,
	))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// broadcast syndicate id
	req.User.SyndicateID = player.SyndicateID

	err = player.L.LoadRole(gamedb.StdConn, true, player, nil)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Unable to convert faction, contact support or try again.")
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s", player.ID), server.HubKeyUserSubscribe, server.PlayerFromBoiler(player))

	// update active player list
	if fap, ok := pc.API.FactionActivePlayers[player.FactionID.String]; ok {
		fap.activePlayerUpdate(player)
	}

	return helpers.EncodeJSON(w, struct {
		IsSuccess bool `json:"is_success"`
	}{
		IsSuccess: true,
	})
}

type UserEnlistFactionRequest struct {
	UserID    server.UserID    `json:"user_id"`
	FactionID server.FactionID `json:"faction_id"`
}

func (pc *PassportWebhookController) UserEnlistFaction(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &UserEnlistFactionRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if req.FactionID.IsNil() {
		return http.StatusBadRequest, terror.Error(err, "Faction id is required")
	}

	// get player
	player, err := boiler.Players(boiler.PlayerWhere.ID.EQ(req.UserID.String())).One(gamedb.StdConn)
	if err != nil {
		return http.StatusBadRequest, err
	}

	player.FactionID = null.StringFrom(req.FactionID.String())

	// update player faction
	_, err = player.Update(gamedb.StdConn, boil.Whitelist(
		boiler.PlayerColumns.FactionID,
	))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// give user default profile avatar images
	err = db.GiveDefaultAvatars(player.ID, player.FactionID.String)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	user := &server.User{
		ID:            server.UserID(uuid.FromStringOrNil(player.ID)),
		Username:      player.Username.String,
		PublicAddress: player.PublicAddress,
		FactionID:     req.FactionID,
		Faction:       &server.Faction{},
		Gid:           player.Gid,
		SyndicateID:   player.SyndicateID,
	}

	faction, err := boiler.FindFaction(gamedb.StdConn, req.FactionID.String())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Unable to find faction from db, contact support or try again.")
	}

	if user.Faction == nil {
		user.Faction = &server.Faction{}
	}

	err = user.Faction.SetFromBoilerFaction(faction)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Unable to convert faction, contact support or try again.")
	}

	err = player.L.LoadRole(gamedb.StdConn, true, player, nil)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Unable to convert faction, contact support or try again.")
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s", player.ID), server.HubKeyUserSubscribe, server.PlayerFromBoiler(player))

	return helpers.EncodeJSON(w, struct {
		IsSuccess bool `json:"is_success"`
	}{
		IsSuccess: true,
	})
}

type WarMachineJoinRequest struct {
	WarMachineMetadata *server.WarMachineMetadata `json:"warMachineMetadata"`
	NeedInsured        bool                       `json:"needInsured"`
}

type WarMachineJoinResp struct {
	Position       *int            `json:"position"`
	ContractReward decimal.Decimal `json:"contractReward"`
}

type UserStatGetRequest struct {
	UserID server.UserID `json:"user_id"`
}

func (pc *PassportWebhookController) UserStatGet(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &UserStatGetRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if req.UserID.IsNil() {
		return http.StatusBadRequest, terror.Error(terror.ErrInvalidInput, "User id is required")
	}

	userStat, err := db.UserStatsGet(req.UserID.String())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get user stat")
	}

	if userStat == nil {
		// build an empty user stat if there is no user stat in db
		userStat = &server.UserStat{
			PlayerStat: &boiler.PlayerStat{
				ID: req.UserID.String(),
			},
		}
	}

	return helpers.EncodeJSON(w, userStat)
}

type FactionStatGetRequest struct {
	FactionID server.FactionID `json:"faction_id"`
}

func (pc *PassportWebhookController) FactionStatGet(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &FactionStatGetRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if req.FactionID.IsNil() {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("faction id is empty"), "Faction id is required")
	}

	fs, err := boiler.FactionStats(boiler.FactionStatWhere.ID.EQ(req.FactionID.String())).One(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return helpers.EncodeJSON(w, fs)
}

type WarMachineQueuePositionRequest struct {
	FactionID server.FactionID `json:"factionID"`
	AssetHash string           `json:"assethash"`
}

type FactionQueueCostGetRequest struct {
	FactionID server.FactionID `json:"factionID"`
}

var whitelistedAddresses = []common.Address{

	// ninja staff
	common.HexToAddress("0xB227842C1742399C92a00AE1c6bb82c638Bf0f68"),
	common.HexToAddress("0xC1cEf962d33F408289e6A930608Ce3BF6479303c"),
	common.HexToAddress("0xdA3167Da376244108c95cAA5d415d2f938CeEB69"),
	common.HexToAddress("0x7D2A2c3443c8174b9F1eeA4549f1deDf2aa8df7F"),
	common.HexToAddress("0x7ED9bD0454108b698ca1560c5565a7f1C8A5bBcE"),
	common.HexToAddress("0x2d1889768243b97e10fd13a2ba9966CAd7BB5866"),
	common.HexToAddress("0x30eE46CdAACCbB8cAD77B815B8A2F36E88Ef0884"),
	common.HexToAddress("0x4115F014C02E17D886BF3eAF50bf213E6aD56EC4"),
	common.HexToAddress("0x3e05174355B9346716F42b2faf9C021e6B9A0412"),
	common.HexToAddress("0xa6339B412df4c88Faf5e380d695C32FF921De4A3"),
	common.HexToAddress("0xeae4020c0c31A791f394b4c64f38ceEd5A02e83e"),
	common.HexToAddress("0x03732540b4F1E5c361decfB9034d8B3f92E46513"),
	common.HexToAddress("0xE990358f7B97201381115e8961aB312249C606db"),
	common.HexToAddress("0x66eB9EF25FB4984464Dc99aFdc1F05ca851F6a84"),
	common.HexToAddress("0x90FE82c9E2823E443f75718155B1c18E84E5E5d1"),
	common.HexToAddress("0x4493e9Ada6FCEC00CDb9C31dCB2fdb60aC792A5b"),
	common.HexToAddress("0x06ead77080393e99E59ae0820b30797248d816d1"),
	common.HexToAddress("0xF63efEcCe28e2Df13E6BE4aA968e623A31c3dF74"),
	common.HexToAddress("0xE87c72e681b09376095AE7Ef9A316C2325ad1796"),
	common.HexToAddress("0xC4377e4DbdEAb7D66C6C102AD5a3156b0764B5e0"),
	common.HexToAddress("0xd3fC0F690097D0241A906735884eaC7A353DacF9"),
	common.HexToAddress("0xec4BB0ea59C5F46113c2C921f99f554A037235C0"),
	common.HexToAddress("0xda96F74BA3eECd26f5E438199d176D07E164E46E"),
	common.HexToAddress("0x636241FeC40AFf3A3F294d8C89c9d47FDcfd107F"),
	common.HexToAddress("0xa242E360F8E98f274ecB2cb705D8624d9d14fb5F"),
	common.HexToAddress("0xc5B46d1e593DBb4383fEd5e8BDf83E61c2711148"),
	common.HexToAddress("0xf9c90E7015748EBE4199f7F345EbA3e2544D51C5"),
	common.HexToAddress("0xAaBa501404A72539752ac93E33A042ac0bB1086A"),
	common.HexToAddress("0x8B40B414Ac058A6fd882c8548ca214D3C78BAe15"),
	common.HexToAddress("0x8085364bA91B1888d922aAA5b28e1B893c91e565"),
	common.HexToAddress("0x5e718C528CCa3B98504671566F9b5469C36aD36f"),
	common.HexToAddress("0x03469FDba3E9f4880E8e9dD7b74d61851afc02f3"),
	common.HexToAddress("0xE2b7AE0b026817e38E29c03c3F57bc697A2Cf21B"),
	common.HexToAddress("0xFa79b76602E644deBD585e254e7A0ea9271Da7f7"),
	common.HexToAddress("0x8080833a93bD3F69A1972452B03dfb338Bef425e"),
	common.HexToAddress("0x3e46B1a261616eb88C6e39B680065451B44Cd600"),
	common.HexToAddress("0x79d4b9CFf3C046F07e9270CE9948b961B3245999"),
	common.HexToAddress("0xdBa43A5434B9BeEF4f473E0e5aEb3B172C7461EC"),
	common.HexToAddress("0x51627A7e67b86decf28a50a9207060f634D6C6d4"),
	// this one is Reece ⌄⌄⌄⌄
	common.HexToAddress("0xEeDBF8aB0D5e20dF93F1539A6b1c18A804335d4B"),
	// this one is Reece ^^^^
	common.HexToAddress("0x3Ca6425be53a9B9cA9650eB8a8B454f455781333"),
	common.HexToAddress("0xEAA5693a4E3cA53A74687440db2E55773b2E3F7d"),
	common.HexToAddress("0x77F9e1c5a2e99Bf0F6129cDb79bAB7a7BeFE879f"),
	common.HexToAddress("0x4F99ca8cA1328C6F44242f7b7333f3637956f046"),
	common.HexToAddress("0x049347822a1535b3159726df32fe264513d9D68E"),
	common.HexToAddress("0xCaC57750dAC844408e5dA080441E09b6619Fb6aA"),
	common.HexToAddress("0x03A2e4336b8de92e2F1Fc9c259A8f574a1Fa7405"),
	// QA
	common.HexToAddress("0x1723997C86Bd0AcAaCEa72531fa5f63eEf7D1B70"),
	common.HexToAddress("0x0ab8FEA85F560FF3C54B68D88479D0fE15fEAAF9"),
	// whitelisted player
	common.HexToAddress("0xE2b7AE0b026817e38E29c03c3F57bc697A2Cf21B"),
	common.HexToAddress("0x1C809993d33e5ecE03330996542536861ED8fb2a"),
	common.HexToAddress("0x6B2E3c751428A181345235074B85D5F952922f8f"),
	common.HexToAddress("0xc622650576F08d9B9f4E1D4C098D69940503Bdb4"),
	common.HexToAddress("0x8cf3BF4a523DB74b6A639CE00E932D97d10E645F"),
	common.HexToAddress("0xEB7Ee71d02Cb518C28f67241b214693ceE4d7867"),
	common.HexToAddress("0xB79F204678801Ea6A10e394b6ed2Baa89737fa38"),
	common.HexToAddress("0x5B190Eb2B2E7dF57a7502945f5E9AEB9FBc27f5c"),
	common.HexToAddress("0x1da05DE4bBCb00F78E72cE1F3cAb17D806cB023f"),
	common.HexToAddress("0x3f0a779FA76D32779b34f3e48f7f4458bbAab001"),
	common.HexToAddress("0x2E345bc15779ce08944195912fd759ef9ddCE9B5"),
	common.HexToAddress("0xf9321f000fd9D25B09894a33D36618d3EA6037C7"),
	common.HexToAddress("0x14b38688eb600B74c27B1E36C8d9d5e8E677eF4d"),
	common.HexToAddress("0xbB2bBA7202c5C85ac9D1F0942d867ec2BE3A303e"),
	common.HexToAddress("0x48A9d56C32C282a8aeB0fC49b702a010C4eBF765"),
	common.HexToAddress("0x3aEC72ca97AAdbfac9BEb8705412CdD3aDc2cf23"),
	common.HexToAddress("0xC59cC9C224424F37D92a69B803E75798bd225E17"),
	common.HexToAddress("0x2793Aa3C7C81Ccd1D7F8480DE2Bd6501E59f75Cb"),
	common.HexToAddress("0xbf61B6DC47A441fbE2B55522DF2dcF34082BE0d9"),
	common.HexToAddress("0xC66A54E60A2672cF9232Ff75E98F78e68A0e16F1"),
	common.HexToAddress("0x85d818eC494f42b73Ff96087581554Fe924Acd66"),
	common.HexToAddress("0xf6d5832c1004b423e1008fd43fd4fb9917023182"),
	common.HexToAddress("0xCc95AC87344827d48b7D96CdfEb3d4a5bdd2E9F9"),
	common.HexToAddress("0x58A084dB3210330910779f3779d8156932a9d6ed"),
	common.HexToAddress("0x0E52A72A9F0d040ED9cc726cC282254272A26927"),
	common.HexToAddress("0x44aa1b7990B36E2dAeC0525cBaaC2f6aA9ec64B1"),
	common.HexToAddress("0xDf8282d36808475C2D213CAFc66a5EE53d73516b"),
	common.HexToAddress("0x850138737C60eF58Afb231FFeB5043c2eb532708"),
	common.HexToAddress("0xFBEF795CB3Dc8705a3E6E9AC92455322E931645B"),
	common.HexToAddress("0xfE1668F3572A738D584957813e6a805e125807be"),
	common.HexToAddress("0x80191032fB4d309501d2EBc09a1A7d7F2941C8C1"),
	common.HexToAddress("0xD0a095C5281e0B8554257918aecfD90A39A2dF9e"),
	common.HexToAddress("0x47EeaA74eF36094bBbD757840Dbda849459568d3"),
	common.HexToAddress("0x3F291f6d31ca58f131fDe2F59aCEB60Ea07A5Ce3"),
	common.HexToAddress("0x5c4138812A05575C927C414cdc8CA7bf8457825D"),
	common.HexToAddress("0x28675EbF67469BBD3ae4FB4C3E01dd880b31c183"),
	common.HexToAddress("0x60C14ad225624Fc9762ec7B588B1EaCaEf43Ef50"),
	common.HexToAddress("0xbF2BB355392846fb52e27af343c81c0c6dc8B27a"),
	common.HexToAddress("0x5b062860935914F1fE1203731E6473F382918DC4"),
	common.HexToAddress("0x02516e8308a1d0c8Be14220296307E207d1e5A99"),
	common.HexToAddress("0xAe6e2E99DEF43a7e7B0a94E5198F30B18E3B7858"),
	common.HexToAddress("0xF9E3a03373bd907A78435382dA2690730Bc1B87c"),
	common.HexToAddress("0x9eaee4cb4bcbd5fb8b3cbcd62cee5f6451cf082f"),
	common.HexToAddress("0x4263f9C484B931E863f9A01cee476053A49DC1e7"),
	common.HexToAddress("0x2237726cED515A5330bdAa6f77355964EA039624"),
	common.HexToAddress("0xA1F880072F6E6145CAF95843799510aB10578547"),
	common.HexToAddress("0x0f0e174d080e08c2749a2aced6d3b9e977282f8b"),
	common.HexToAddress("0x8E856Fd170d44580064D0AACB2B3B6E6Ae331EF5"),
	common.HexToAddress("0x525D92Ee9fF660e7DfC781A9c35497B1CAaE19Fc"),
	common.HexToAddress("0x5D6984E9D21Fe1F75bEEC2EBdc0A2E066183855d"),
	common.HexToAddress("0x969Af35d75C10fd5d0B57E322b415697E06cfa90"),
	common.HexToAddress("0x48A9d56C32C282a8aeB0fC49b702a010C4eBF765"),
	common.HexToAddress("0xBa0FbD09ECA2e2a6Ad79D2e2F9f5389c667Efb86"),
	common.HexToAddress("0x010FE8CCC138D35aE69f197b23fa9Ea2Fe129FFd"),
	common.HexToAddress("0x035A71B55c902aFa341f81eC3Bc6f4A4e4E3dc30"),
	common.HexToAddress("0xCf33e657eB463fE01EAc42dF9234C2f3936811AE"),
	common.HexToAddress("0x6496039a7Fe183156b3a90652A794Ee9C2FbE7bd"),
	common.HexToAddress("0xB8cA85a8C25AbE6C184055830e10823588da1E6f"),
	common.HexToAddress("0x0e423b4a193004340bcee5a7fc4268f054bbf774"),
	common.HexToAddress("0x976C1E455b75f57931f3019Dc3D0E600979dAc43"),
	common.HexToAddress("0x06Ef623E6C10e397A0F2dFfa8c982125328e398c"),
	common.HexToAddress("0x122bD794009DEF11216Accd895bf3bcD0Da51008"),
	common.HexToAddress("0xA7Fc97E08340efdf583A0437c52B525cc9f56138"),
	common.HexToAddress("0x81bc6403665D71f65cE2bA359BE15B98a215675a"),
	common.HexToAddress("0x26Ba4FC26fB5048adcA3403e5B329272F71985A5"),
	common.HexToAddress("0xDBBc9be51BD9C5Cb90D8D3E74b23e1D5114E7387"),
	common.HexToAddress("0x603a4f72e004f24a5c26667c767995c5b14a37f1"),
	common.HexToAddress("0x1E29077a5B9F29F80088179E462dEd5C49F301bC"),
	common.HexToAddress("0xf64ba532851619cD28924D933446022D349d01E5"),
	common.HexToAddress("0x0e6BE047eee4677869A138D6B5a1E87bf33e3C29"),
	common.HexToAddress("0x616978005a7940d03d7e3C472810f32ef0Eb7a24"),
	common.HexToAddress("0xDFA1e36e88F7Ac449425BA8B1095f033E606E9C2"),
	common.HexToAddress("0xdf427B2aa315E5E9991F249d1664675Ea7EBb9Ac"),
	common.HexToAddress("0x43cb4bc9551966D553436900BA4d835F2a7163A7"),
	common.HexToAddress("0x99Bb438056Fb4075CCCd476FBb613154370c2F86"),
	common.HexToAddress("0x33C9Ae0FA7ddFe4278C9CfE5ad09cFCA061FF246"),
	common.HexToAddress("0xBEAa9d6d1248b7C34817A3018EDad256dDA4a762"),
	common.HexToAddress("0x9A2b8Ce6eca92287cB9E323447CAF54f311a7c93"),
	common.HexToAddress("0xF61887e20Eb20E2a731905FC5Ce3d22C9604653b"),
	common.HexToAddress("0xDef01332c7F8305dEE80B6d48657CC8db7cE9ca5"),
	common.HexToAddress("0x785aFB5b97ea4158DC026D9C2a711dF0D723ef8D"),
	common.HexToAddress("0x301afFe4Ca5f4D6A1BcB36Ca9c45c3fDA917777b"),
	common.HexToAddress("0x2549A59780DDE0F69326E27FE41741A9c39B428a"),
	common.HexToAddress("0xaB15BAb6293d1F8aBfa36650A2b81C70B7182879"),
	common.HexToAddress("0xAa21228C52F7623cc147dc326179D5C6a2aE4ff5"),
	common.HexToAddress("0xc348Db10163E9565BEF864591582bB3dAec25857"),
	common.HexToAddress("0xE216Ff69D164f6551ec36BDfe6FB57e45833D6f4"),
	common.HexToAddress("0x2288a4E1E84459FD55D255F5dfe47FBA2ef10aA7"),
	common.HexToAddress("0x44f3d981488dce5a07D20eA7670bD3614c6da153"),
	common.HexToAddress("0x32ebddb207622d47746a7d0caa18c17946474830"),
	common.HexToAddress("0x3494454C2B2F961b6Fe4Bc917Ce0265cC2ed6799"),
	common.HexToAddress("0xbBdB47F3B286aFC84884d75475489c5Ed74B00fF"),
	common.HexToAddress("0xFf17234FA1AFa7692eD3ef9884786f11425807b2"),
	common.HexToAddress("0x158D67aF0AE1B02E4EFa150C5e77F60893Edd769"),
	common.HexToAddress("0xf0A7664B766eF63371e97d57Ae1895ae1Cb0F726"),
	common.HexToAddress("0x5c87ac94848107e99a56dc5b55aa26969439201d"),
	common.HexToAddress("0x847D979EBa43e3436F6DA5a6A0d24ba586021510"),
	common.HexToAddress("0x1B43dBC1DDdc96A6546e92683786DC57e601276B"),
	common.HexToAddress("0xE848745D0C215EB4643a856c97272E1cD7fB3Ff2"),
	common.HexToAddress("0x7d03682aba72f841F70314B43498548e5c5fBC81"),
	common.HexToAddress("0xe4fF7Ebb4Cf5889492fBffa6ac7d57cd7BbC3d0E"),
	common.HexToAddress("0x824BBe8Bd445e2F7f1bc7292e1807411A551f288"),
	common.HexToAddress("0xD98942f2D07890591dc7Aaf78a2C05b2355839BB"),
	common.HexToAddress("0xEc227f2b29f0bc50c7a06ed08882a1367bA4aF48"),
	common.HexToAddress("0x68650C3DDd88bf43557C6CB7fea97FB3366a39fF"),
	common.HexToAddress("0x0f7532cddc83481487BEbDe7af2a4C2Facd97e27"),
	common.HexToAddress("0xba480edf393630d0c9f2a20f6ab072eb2584ec4a"),
	common.HexToAddress("0xe545A26DD6aCbC146b80981BFC969d5d47959C0b"),
	common.HexToAddress("0x8c448Fa410c3d67D80AC985Dcf1f42147803549D"),
	common.HexToAddress("0xcB873238Eb6fd2B857c1379ADcB107082cffBc4F"),
	common.HexToAddress("0x4e1A7F0446aFA442b8F88bf88f6b9139c9b1266C"),
	common.HexToAddress("0x8AB91eF74c94AdB0558952d2B7A8824D13dbB0F7"),
	common.HexToAddress("0xc27418d92e4614fdb8d094ef89a009a776ce1bde"),
	common.HexToAddress("0xCe3ddf9436bFCC3D9bdb1810B88F07EE84da5616"),
	common.HexToAddress("0x08b07a54FA332bC067B39507137829ad9B315489"),
	common.HexToAddress("0xDCc4917DDb702c120B245E819F0df742043E5AdA"),
	common.HexToAddress("0x366Cbe37Db54D4d72108fE827A9F8beB16A00D51"),
	common.HexToAddress("0xcA4b064b97A072fB23535677fD22E52f74390343"),
	common.HexToAddress("0xd305e634Ab5F283018D243cC7114999Df66efF2B"),
	common.HexToAddress("0x7164e91E07B7BA30ef0dDF0eD51e8F65999D73ba"),
	common.HexToAddress("0xc1EaD6321541B3e37A40BcF898e7E5C3fCA9f2D0"),
	common.HexToAddress("0xb538dd31ddd4533E8865e7A821d58b9C6CA780fB"),
	common.HexToAddress("0xd6f12F9c4733d471d8f82d3AB76bB40d50caccd5"),
	common.HexToAddress("0x5fEC2F30c7d74C70e57cD75Ce7239CF3EF61babA"),
	common.HexToAddress("0x21599bD8de3C04a3db136952E9CEcCA6E9096b6f"),
	common.HexToAddress("0xDfd09defc223228CC85240AC832B3b149e036eD7"),
	common.HexToAddress("0x9E9c0A5cB6F4f63F6ED7b7C00e4821f8f95C0510"),
	common.HexToAddress("0x218C609244494F2bcDae80aCFD811775c4EA34B9"),
	common.HexToAddress("0x90ac2CFf78235623bD534033c9FC9d2D6AB0ad39"),
	common.HexToAddress("0xd61b54ec89e3e3A5113CF1378AC3637949cc93C2"),
	common.HexToAddress("0x78AD24689ED538425cF0BF78e10De80C6B9D6aC8"),
	common.HexToAddress("0xAeA873c12b6fA72E5104f4b3145ee6B3C915ca6C"),
	common.HexToAddress("0xe1284c4F84ab2D7bb4D8f1569ca0C9037cd17f13"),
	common.HexToAddress("0x1B760F7F2e8C2F2346E8410642aF72a6Dfd6BA2a"),
	common.HexToAddress("0xF5E700E2A1F1Bc18C7bbD16c7Bf23CC4765Fb610"),
	common.HexToAddress("0x33113376710D44f4A878743341C5Eb942484C20e"),
	common.HexToAddress("0xDb37Fc0F86797634f1014AE5006794431C02691C"),
	common.HexToAddress("0x14557B7E2055E84F4F8C762cEFCb8Cc13c537259"),
	common.HexToAddress("0x50f0B770aDb1FD8B9c6d65e4f45C5fc876DfF544"),
	common.HexToAddress("0x04302BdEB72a36e418b0b91014284CF5105632dA"),
	common.HexToAddress("0x26E92E14A2A6Cb058F4DeDB6a3340c35a7a8cDbe"),
	common.HexToAddress("0x7Af4Fa9dba69C380f38D1135e36c157BB9c9d894"),
	common.HexToAddress("0xB48a8F10242AeC63206BA9F77988eb19F3863B88"),
	common.HexToAddress("0xDa331BC04245D6A677Aa96b62e7E06be3c928A55"),
	common.HexToAddress("0xdce1cd744b39c902e0c2EAA28328Fdd72AbE748c"),
	common.HexToAddress("0x5783592A078b9502e76Ac57b91Da85C151FBBD1c"),
	common.HexToAddress("0xA541c6Aad52916afa701422a1DE3955c40DFcE06"),
	common.HexToAddress("0xb8A8f6ec7F565e71e934bC1a3529a0fAFB77880F"),
	common.HexToAddress("0xAC6dDe31b81fB199eBfFaB674dE7CB8A50023771"),
	common.HexToAddress("0x5e602b771bB9555E7077332Ee625643DA821e021"),
	common.HexToAddress("0x7AAf87F6F1F9105e6f32bDeE706329Cc1639B58e"),
	common.HexToAddress("0xA5bA99503fE74e7d435306c5816bBF91fCB317F1"),
	common.HexToAddress("0xA559263D14d08dc495A083B69FE0FBd1968B8f74"),
	common.HexToAddress("0x208280e826d0195aBae7177C9408AD795465e364"),
	common.HexToAddress("0x58fDE9484303ad05752520424b29371557e4a4A5"),
	common.HexToAddress("0x17A0dc922ee02DBAAB2329eFB58f79ac65F347fd"),
	common.HexToAddress("0xFD3B74C74fE08A6Bc39FEcE3DEF182008c270c5d"),
	common.HexToAddress("0x0A404B31F5f7dfd5B14d50b33a506Cf64aF03eB3"),
	common.HexToAddress("0xc820709e01470282fCd7AD168210f2feb2C41837"),
	common.HexToAddress("0x48E14ecDb41a298BBfEe72e643CEa0ca485Af38f"),
	common.HexToAddress("0x407D46DAB64Ac1698021e29cA0a21B1B0bb7f4bD"),
	common.HexToAddress("0x37998495c09662E26b021Cb29c6B7859E97Cdc90"),
	common.HexToAddress("0x1052B9F4A8FAb42d1562aBB2df8aA04Fbd006572"),
	common.HexToAddress("0x98A1B42080aF83b752f37755489F6335AF4145f2"),
	common.HexToAddress("0x52eF9EEFDac780A660D7fae1C04Fb665be3aa685"),
	common.HexToAddress("0x31AB6B3c2c52c0456Da584895d564e169BD11AC2"),
	common.HexToAddress("0x6f46C12a80dD29f5165d07e98b0B5D948D94BAe2"),
	common.HexToAddress("0x926ce9fbb32021115c0b18830a74b906837f10f4"),
	common.HexToAddress("0xDA132eCce452f4030C25329e739e8708eE2E6660"),
	common.HexToAddress("0x8a3286fd1ac65b215d33CC97616e1BC9508dB431"),
	common.HexToAddress("0x1D4F381F33a4A18c363a00a71CD0B81aC5b9f202"),
	common.HexToAddress("0x66f6579d72f5399e5782D29Ecb1EE490aFd5Df5C"),
	common.HexToAddress("0xB44AAeA857Be7246320514f660431fbD0680A8Ad"),
	common.HexToAddress("0xE2d98e889Df20B54C54A86C2A1Af88169714552F"),
	common.HexToAddress("0xd458f2a2d92f02b4a86f1690810e70844c5be02a"),
	common.HexToAddress("0x221B837239713d88539B5b806f8573c909B2ed57"),
	common.HexToAddress("0x953cD496c2371aCB11eBBA340fF548C4F50f2f02"),
	common.HexToAddress("0xBa034dCfd0735aE3C5DAd25Ccb8E0E25BBf28788"),
	common.HexToAddress("0x406E0d885F3fbEC20d4EecbEd8AE4370C1fADfAB"),
	common.HexToAddress("0x3adA4368e3f86B44E921DEA4c13CC73634C7fEFa"),
	common.HexToAddress("0xD0FbD1d387D873f3FCA827C9a5Eb0fFd759d88DD"),
	common.HexToAddress("0xA662bBfbb8FDd1Bf929Bb3a26B8f780123826084"),
	common.HexToAddress("0xA3E73Cc5A015eBbBf8a748dab77c41972eA350C4"),
	common.HexToAddress("0xd91C81e155F547C9B9Ab14eD240932F6B5e2c3b0"),
	common.HexToAddress("0x0C6146B4456792fc7e9D83aED4768e535CB9B766"),
	common.HexToAddress("0x12702da849f349d99017a7e2228599f51519b864"),
	common.HexToAddress("0x2A9Ee384F6A3298d0Da1ad0bd383441B1b634F87"),
	common.HexToAddress("0x249204676Ae79d6F6d2211dBBE91D961b15fEf71"),
	common.HexToAddress("0xd061a5642A9A9046a774730a2d790CdEc698450F"),
	common.HexToAddress("0x564fFfeCc261E1B4EE362674d53A95989E3e13b4"),
	common.HexToAddress("0x005d58118836cE18F120c3990c4e0aF2DbB06331"),
	common.HexToAddress("0x944b57700eca9b55319E4D43ECe13BaFA032D62f"),
	common.HexToAddress("0xe44745b6E21B5E0CFD3e004E8eA0C213e6b7A030"),
	common.HexToAddress("0xfE1668F3572A738D584957813e6a805e125807be"),
	common.HexToAddress("0x19aadF5652f780D1c065572203152bc675998685"),
	common.HexToAddress("0xd28fF0E226b80908ad535e0d4E49D18Dae952076"),
	common.HexToAddress("0x82aab7669552b2a0ca0463b52635b17420a2599f"),
	common.HexToAddress("0x82288d961Efd6A8bc068aF3eFEd446E0aFE199Af"),
	common.HexToAddress("0x49F9D4594877Cc733E9F7E66B533d891DE291590"),
	common.HexToAddress("0x26b85eC4E81bc1621c30371993a8EC8CdE0C1FB6"),
	common.HexToAddress("0xd5d23311ba1A5769cD29AB208f75a4042626AAaA"),
	common.HexToAddress("0x8085364bA91B1888d922aAA5b28e1B893c91e565"),
	common.HexToAddress("0x44262Bb3dfe9B3461Abf867fDB1CCFc1C0a2Ab34"),
	common.HexToAddress("0x493f5Bd535630F80C5e4B63f029da64826810194"),
	common.HexToAddress("0x2a7f3933e3e12350819ad23ffb067ae724e90854"),
	common.HexToAddress("0xf4EC5ADA81e9e954A7F30c09346affd0ee56c9fc"),
	common.HexToAddress("0x602e4B163Bb007D5e22DF28d3456129DA2cBBFb9"),
	common.HexToAddress("0xC24a924FC4e57c6a12cA96a2C84179a16034d8E9"),
	common.HexToAddress("0xD2779012E62c8a40437A84E3A2102D6AEA2B94B6"),
	common.HexToAddress("0x2c9388a50955a823f29D0554B39E8593cF8cD284"),
	common.HexToAddress("0x59530f01f4f3148c327f95e03963c62fbD1fAA47"),
	common.HexToAddress("0x4E3586902710fa99181B1ea91daBE771c343CB06"),
	common.HexToAddress("0xFDd16FF32b276B68B42CF68580A8BaFb36773eA9"),
	common.HexToAddress("0x0019E1B74a0bF8127515D1f72D0b3ADb619dA6E3"),
	common.HexToAddress("0x0a46DFa93f45eDFCb0ba928CEDEaeF463b267C34"),
	common.HexToAddress("0x5547e6841cFb4bBC9B2386db0f9212bE32559f28"),
	common.HexToAddress("0x06Ad668aa9FE73fe83C146494BeA28B38350eDca"),
	common.HexToAddress("0x26C19F9a7cf073f280F6d1F50dbD464704DB2cdd"),
	common.HexToAddress("0x3eac640dCB934FFe18B7239Ba34618A81692383F"),
	common.HexToAddress("0xef30c130f5061A171567169AdE6c8D2FBeE6C3D6"),
	common.HexToAddress("0xf27cFB91AA68402A794a85BE6863C502ac5bCefe"),
	common.HexToAddress("0x1156b08881B4b1c759305bD7fe1D6E13DbFC67Ad"),
	common.HexToAddress("0x901857258a1CC55E67C253f3B783373983a4D42F"),
	common.HexToAddress("0x4538eDd7612F11B0BA43e346f5b52746F23A13d9"),
	common.HexToAddress("0x3C528c054a245F39992cCeEF3924A069dD3fb6c1"),
	common.HexToAddress("0x404A18541c4f7E5d7442AaBf5aF3DDFD238D39da"),
	common.HexToAddress("0x0166219c360c987A7C82e1EEfb7437268321A853"),
	common.HexToAddress("0x33113376710D44f4A878743341C5Eb942484C20e"),
	common.HexToAddress("0x5d619c764F9D3EB62AB4bfEd2d87B892E286D24c"),
	common.HexToAddress("0xFE29aeAedA2c4939D0561A38B47a43a12D559aDB"),
	common.HexToAddress("0xbd2DEFa576D9fC93e31706aD7F994bC473D024F5"),
	common.HexToAddress("0xc5bb46e33b7b8263415e6b4c998c221e5af6738e"),
	common.HexToAddress("0x4D37D6b4aC017dB4E57E20cB2d51e0481Dd852C4"),
	common.HexToAddress("0x42e24ff7ad60ac90b10d320ab1cb75779a8e6054"),
	common.HexToAddress("0x60504C8342838d4d519957D3FEd5181695b553DE"),
	common.HexToAddress("0x43ae30820DB7D63F1dB8Ac70FEd2028dDf7CD001"),
	common.HexToAddress("0xed629aFf431ccd5cAca395203156371aCde2f272"),
	common.HexToAddress("0xb9f1A5d1437E3E2365f5bc5F097666B8778915A8"),
	common.HexToAddress("0x0d271fE59BccA1De188456e38e716D0e080d911E"),
	common.HexToAddress("0xE8B15F1e49E3580b21B886ded872Bd6b76f6ba3E"),
	common.HexToAddress("0x1B294d9b4c974CF97d7A129f14EC2005fF88D554"),
	common.HexToAddress("0xEBC3e73720113bE3cfDead098E9f782fF6840b36"),
	common.HexToAddress("0x0019E1B74a0bF8127515D1f72D0b3ADb619dA6E3"),
	common.HexToAddress("0xD29F440051F2D2666D74f1101E67928babFc917b"),
	common.HexToAddress("0x442C982b09F77CDE9161E4CE4E69aDB544612F8b"),
	common.HexToAddress("0x14717DE0a0E507836A44539f76d095755991833A"),
	common.HexToAddress("0xb6E4119BE533fc25c9791f980Aab068895Fa6045"),
	common.HexToAddress("0x86Bd28AAc997Ef48d871a409a3Bb84ec0310f52F"),
	common.HexToAddress("0x94f6ec5539A2649F3a69c8be7e7Ba44bDa2e683f"),
	common.HexToAddress("0x4115F014C02E17D886BF3eAF50bf213E6aD56EC4"),
	common.HexToAddress("0xE1E02Ab684318b128E4D3Fc1b4a4556dB0f7f408"),
	common.HexToAddress("0xea7dDE03961BE9A2D6903a15cFD420Cc6FF4CD66"),
	common.HexToAddress("0x8DAb222EB686a4f0708e892D09fb3558249c916e"),
	common.HexToAddress("0xa431B3FCdF8e42BA42B44B3c582030d0Fac28B08"),
	common.HexToAddress("0xbb8aD5dF6362D736E622C71B6942a71DBcbC537C"),
	common.HexToAddress("0xB3110E3607af8d175212180A3D6Cb4B98De46362"),
	common.HexToAddress("0xaf90CaE076aB4EC57eb5c5eEAedEa9C7821dA1B5"),
	common.HexToAddress("0xb0C17c71548FD3b91aDDE2F154DF739ccD38F959"),
	common.HexToAddress("0x4985dCC3e2b824C25a6700a742154B84DaC3E9EF"),
	common.HexToAddress("0x6D697c4c2dE019A354eeECC2Cd6797bF862468be"),
	common.HexToAddress("0x9BdB41893cd8A999018ac1C39B9Eb31b8A13f035"),
	common.HexToAddress("0xe2a505Bfe8FdbF897310Ad89364cB0e04B5E3D9c"),
	common.HexToAddress("0x9eaee4cb4bcbd5fb8b3cbcd62cee5f6451cf082f"),
	common.HexToAddress("0x9eaee4cb4bcbd5fb8b3cbcd62cee5f6451cf082f"),
	common.HexToAddress("0x2A67528AC463790f9077f16BFad065763E3b1140"),
	common.HexToAddress("0x5C383e973899633C9D669b931B180c2780A30696"),
	common.HexToAddress("0x354a06c3280D0bA61841Ba86D36048190650748B"),
	common.HexToAddress("0x0B2aE71C3aBA1f72b9a1d3e263f0B05d2eCD09B6"),
	common.HexToAddress("0x77cF7F0f9875Bb0BAdF9Ac9407398F1Ddc048499"),
	common.HexToAddress("0x77cF7F0f9875Bb0BAdF9Ac9407398F1Ddc048499"),
	common.HexToAddress("0xC0F54882c43C121b6791BD831D885A45f5080712"),
	common.HexToAddress("0x5c4E7ace29fA42342a618C46497fB886626F4A0A"),
	common.HexToAddress("0x35c14836c2542a4590448b1114cc0f5C6067f11E"),
	common.HexToAddress("0x7298f78afb070EeFBce2b4dd13A84c137751D7d1"),
	common.HexToAddress("0xb6d952190729c9Eb80b34C5CE5dFd5C3921cb9e9"),
	common.HexToAddress("0x78E49baC0DeCA8b5d7beD026bbEE685975eDB834"),
	common.HexToAddress("0x705E290F51A6614BE78D4d321B582309fd930E97"),
	common.HexToAddress("0x0365e7423E42948a12486E3206fCfBa0e9dcAad3"),
	common.HexToAddress("0xA8543E962EBd40ea46Bf2D97E5fF7BDc8893baAa"),
	common.HexToAddress("0x15bb9e3D7A926928D9c7A73C8896361D35814d40"),
	common.HexToAddress("0x0af1e5f724d4872b0615afa7546e89Ee2E21AeBF"),
	common.HexToAddress("0x3324c0aF95EEF8c493D80446fc4186ac84443399"),
	common.HexToAddress("0xA7c41143c20559eC06af569098c141b0e1DC8Fd0"),
	common.HexToAddress("0xaCaed4ff22E6230f0e8E1A93C2eDee0D725E49A2"),
	common.HexToAddress("0xb538dd31ddd4533E8865e7A821d58b9C6CA780fB"),
	common.HexToAddress("0x3C45dEF35A079da4628b918Bd2c7C6D2Ea858236"),
	common.HexToAddress("0x8c89758EB23623bbC0d7a681637f006894c60066"),
	common.HexToAddress("0x34433C06518640F1aAC804eA5a3Dda441950FD43"),
	common.HexToAddress("0x5d619c764F9D3EB62AB4bfEd2d87B892E286D24c"),
	common.HexToAddress("0xDFd5Ea1FeeA4e91C524438366F6B6C6B29E8Db02"),
	common.HexToAddress("0x1E1c271789dECA0d28d9b7DC148Bc1adaA557Eea"),
	common.HexToAddress("0x9CF805447b7E2a8BB3Aa6cD6eb310A3d54BC70Bb"),
	common.HexToAddress("0x11b0df837097f81daC2a70D884D9169eeE7e1F85"),
	common.HexToAddress("0x9F50Bb924714bcE0159cC5aCe5A8b0c68f4301a9"),
	common.HexToAddress("0x5E41B244223cc2e6832Aac1f7f770be967bAc27F"),
	common.HexToAddress("0xF61887e20Eb20E2a731905FC5Ce3d22C9604653b"),
	common.HexToAddress("0x5706B32c0Ab5d8Ee799E021026348250401F0F73"),
	common.HexToAddress("0xc37AECf7E38bC9E32FEEfc29EB6d24d554AeD086"),
	common.HexToAddress("0x3e8a1af9ea608e86b9a9e10b74ef4e92ca4b71a3"),
	common.HexToAddress("0xED0A8a8b416C6eD78337A597649Ccf586Dc09A02"),
	common.HexToAddress("0x2Fb4d0919936E32674d4ae3AcC4EAa1745cDeac6"),
	common.HexToAddress("0xD893Cf5A7B6964c7dbFA82dB383Cf1dB5aBa65D7"),
	common.HexToAddress("0xba2ddB9d30de3B652415e13d323e4d1A5328CbCb"),
	common.HexToAddress("0x50E356f40dCD789AF4150bA68B03Ddc4FF0790B7"),
	common.HexToAddress("0xa4792e4d06872801b3893210e13dd7e68c7b4518"),
	common.HexToAddress("0x23545249652E29AC3da2a99CC6BCD3FAcFf8bB5d"),
	common.HexToAddress("0x2B0666F128374Ce8F30d7560bdAF2bc14e079Da8"),
	common.HexToAddress("0x602e4b163bb007d5e22df28d3456129da2cbbfb9"),
	common.HexToAddress("0xD4128925eeB834aD0b7C6b3112328fCad7eDbfcC"),
	common.HexToAddress("0x05333F8D7c500f313Ad9dd83b367253bF56333ad"),
	common.HexToAddress("0xD288F2F3F02a5b68f1B20777566596006dc893dC"),
	common.HexToAddress("0xDE0Db07A0B54cDbB1f4F8d34309aaE5c3bC7C68d"),
	common.HexToAddress("0xE1124FfA1df2c8eeC196eEfa07Ab81db48e28Add"),
	common.HexToAddress("0xfD6797cfD96Ea1401408E482a3af916b45EF26bc"),
	common.HexToAddress("0x47EeaA74eF36094bBbD757840Dbda849459568d3"),
	common.HexToAddress("0x916EcB606F0a93E40bB7C6c21a82B54650408243"),
	common.HexToAddress("0xC9FF197bd15dbC0dAC57aaC903Aa2bA634Dc60d8"),
	common.HexToAddress("0x2E8f1e23746375b680fC013403Ca078A70d1fF15"),
	common.HexToAddress("0x8112190DA14D9042f5C6792934870C7059981392"),
	common.HexToAddress("0x3C7A481DE53E405515606e9D11EdD789bD38b505"),
	common.HexToAddress("0x1052B9F4A8FAb42d1562aBB2df8aA04Fbd006572"),
	common.HexToAddress("0xbb4a7544c861b7e4ffffb25b47889de78a63d68d"),
	common.HexToAddress("0xd36e4E1c0D02fa462A14Ecb1A0Af123CAe752d09"),
	common.HexToAddress("0xdbd3eF9311b70E4ea09Dc798DBdf6090C2954C08"),
	common.HexToAddress("0x548472Cc3A74401E818604F80F3E99e0A89f3625"),
	common.HexToAddress("0x1588231B8Bc1e051f054496Bef311eeB7e8fB4d6"),
	common.HexToAddress("0xae0a263A8E5Eb1f29801114f1A38840791F5E31c"),
	common.HexToAddress("0xb6DAC350db842A1F2F97481a9128Cf0F37870bFA"),
	common.HexToAddress("0x40995080be52C9516266222D95A188dC742CF6f6"),
	common.HexToAddress("0xc0E1e74544Aa648866f496a23A6D2D25bf7Ad1e8"),
	common.HexToAddress("0xa748bF12eEc5708E8722e9f1c5342a4b2A7E6EdD"),
	common.HexToAddress("0xb217C32955c9BacEddad6204DC7FF509EfED5A6A"),
	common.HexToAddress("0x525fF3b944511deB1B0f4e65950d0Df9bA1482B0"),
	common.HexToAddress("0x27351A3fa32C734E272531f7e9306491Fe881aa8"),
	common.HexToAddress("0x3e7ef63eb946BFbFccD98eEb0CdA43A0B9a0660C"),
	common.HexToAddress("0xba480edf393630d0c9f2a20f6ab072eb2584ec4a"),
	common.HexToAddress("0x953aD7C3d21Ee7a74caE3A0341ddbDB923ae24d6"),
	common.HexToAddress("0x720F4984fb5f15cC4EC0bc1128CDf54D4594bd03"),
	common.HexToAddress("0x406E0d885F3fbEC20d4EecbEd8AE4370C1fADfAB"),
	common.HexToAddress("0xc186cd70379e023EC4be55CC254f4803A77808e0"),
	common.HexToAddress("0x635e41dc964f052f41Bff915BbAFF50DEa57DB47"),
	common.HexToAddress("0xadd21E53777E06D59970994F751751b7302a72c3"),
	common.HexToAddress("0x19B98C0473A1e06caf5E16037437b2db5725841E"),
	common.HexToAddress("0x5C383e973899633C9D669b931B180c2780A30696"),
	common.HexToAddress("0xBDe8c4CeDEB7d2e26A65276168a3c92a065Ce4D6"),
	common.HexToAddress("0x514491F7D867c5c30c2659E836c7409609B14C3e"),
	common.HexToAddress("0x2CfF6fB5a463735D65c29d933CE7f1C45350cb33"),
	common.HexToAddress("0x2e967C1493fa3A8a405a9b8d5cA5E39a6bB5f338"),
	common.HexToAddress("0xC82ba1B579D36Fae835F54eA3a2A83D93a54d446"),
	common.HexToAddress("0x525D92Ee9fF660e7DfC781A9c35497B1CAaE19Fc"),
	common.HexToAddress("0xE918dA9A4987aB2321a2596f05c59d4833d083F0"),
	common.HexToAddress("0x9D8b8e2fd9D1B5a1d143423FDCe85C9b63009169"),
	common.HexToAddress("0xF23C83EfA0Bb64aF1B674356f20A1593C9453966"),
	common.HexToAddress("0x5718f979B454D6cE7ef2aD192F3704BC46b08ea1"),
	common.HexToAddress("0xe057e5404dfb803e54933A57cdeFe39315ef6d38"),
	common.HexToAddress("0xe057e5404dfb803e54933A57cdeFe39315ef6d38"),
	common.HexToAddress("0xe057e5404dfb803e54933a57cdefe39315ef6d38"),
	common.HexToAddress("0x21fd39e600579D9a76a3da7B8aA97d861194a2bF"),
	common.HexToAddress("0x2a9DF41C50bCD31DC59Ef725C95ef88516f59C02"),
	common.HexToAddress("0xa0f73D589cdAF8B8df41f9b2BAf43839DA4d3A21"),
	common.HexToAddress("0x2a9DF41C50bCD31DC59Ef725C95ef88516f59C02"),
	common.HexToAddress("0x2a9DF41C50bCD31DC59Ef725C95ef88516f59C02"),
	common.HexToAddress("0x6474E75122333A666dFAE4fcc2AC461d1a3bD245"),
	common.HexToAddress("0x2a9DF41C50bCD31DC59Ef725C95ef88516f59C02"),
	common.HexToAddress("0x2a9DF41C50bCD31DC59Ef725C95ef88516f59C02"),
	common.HexToAddress("0x035D7Ab350ed549c22a0BCD24412F39391144F41"),
	common.HexToAddress("0xb932Dd9EA91f270eaFD83c8a72c8977869A15a48"),
	common.HexToAddress("0x5Bf973C63688286c15043aF22fb5e75F98D6ae79"),
	common.HexToAddress("0xbC46e7962C4A6940E72e2866d6513F37C3F90b6B"),
	common.HexToAddress("0xE88563dA25BCdF2040ad8955c16C93709597902A"),
	common.HexToAddress("0xB5dC59930384d5101F8fc00D60ffCD659D167F41"),
	common.HexToAddress("0xC4f41654aCa0b55782e6477c1F4B34AF653313fD"),
	common.HexToAddress("0xC12A256dF8870770cC07488E4ea83c92ae1e4e51"),
	common.HexToAddress("0x8CcF71c5B58c55a0bf7265d121e91F45387F0786"),
	common.HexToAddress("0x0c8080d3ac04A84df55Fec08E290F3Ad6625bEd8"),
	common.HexToAddress("0x56Af4E6De4e70029983966c8ED33b9beFAec3159"),
	common.HexToAddress("0xE6c5648Ce030aFB8F7514Acffce7753Ddd8440e5"),
	common.HexToAddress("0x98F5e6295Fe68dcE1aEE28b73fc70dF35db880B9"),
	common.HexToAddress("0xDDA9fFa8199D8868725BbBC27797EE117eEFc55C"),
	common.HexToAddress("0xb0Abd6A7F7591ABB090bf2655337dEF7174fE7b4"),
	common.HexToAddress("0xEB17743993476cE027f4128b986b1c681488d778"),
	common.HexToAddress("0x4931901D7598Fc343CbB1b3908c759Ea60371BBb"),
	common.HexToAddress("0x3Cae6527D63507dd5a019f9a5802Ed3E333C4f56"),
	common.HexToAddress("0xEbc5E1739616031Bb9F7307A5301a36e1F0A7e73"),
	common.HexToAddress("0x06097899b47bfF77D3dBDd6A989410d71c7e4F5B"),
	common.HexToAddress("0x2Fb4d0919936E32674d4ae3AcC4EAa1745cDeac6"),
	common.HexToAddress("0x44DD57562b2845e612AC05b41B60e7105aB36052"),
	common.HexToAddress("0x854928e9a7E589de17b1513c58C35ddee10949Ea"),
	common.HexToAddress("0x406E0d885F3fbEC20d4EecbEd8AE4370C1fADfAB"),
	common.HexToAddress("0x52bA3b6bb5D158386f63ce6a7029a6427ed91589"),
	common.HexToAddress("0xbC46e7962C4A6940E72e2866d6513F37C3F90b6B"),
	common.HexToAddress("0x90155CE7197F245e6BCF7C8d7486883179594Eb2"),
	common.HexToAddress("0x50426f8009ED09450573fbC50322f6d74dF8898D"),
	common.HexToAddress("0x4dac1ecd9F662deAa46549D1e3A404D0cb9C190C"),
	common.HexToAddress("0x30c90faA26781399cE206a87Ae3cB3BEaA6607E1"),
	common.HexToAddress("0xbdd6d256dbac5dD954bB30523919B1F5aB471dB2"),
	common.HexToAddress("0xd4914b3460Fb06DaCf2EaA81482D6C78510dA495"),
	common.HexToAddress("0xbab7b4CD830f713C91718c5552e6ADfCb2Cf200a"),
	common.HexToAddress("0x25ad1017b7bfa9f2587aF2C6ed2650cf9EdCe718"),
	common.HexToAddress("0x6D37fb8174A19b1556aC12F14a4D06088D5318CE"),
	common.HexToAddress("0x041984dbc74a4e8F82a96b03b2f16b8E520CC7ED"),
	common.HexToAddress("0xf06C6eAA599D4d2f844419A8B4dC68FB61018a93"),
	common.HexToAddress("0x57d178CDaA5F0BEe4bc30e5C719F10D6DBD0E62E"),
	common.HexToAddress("0xC4C8ea52eA1252178F8F24237f05F779D8Dea220"),
	common.HexToAddress("0x1BD1136a012312e454CC35F7a80D3b2e4Baf5CFF"),
	common.HexToAddress("0x2F260959Bf5D3ACBAdD912884621A5f906ccc3BB"),
	common.HexToAddress("0x6D8B5Ab0216a8B285afd3811468B18fBb79852d2"),
	common.HexToAddress("0xA4050d47E3435Dc298462d009426C040668F4297"),
	common.HexToAddress("0xF24530D32698787bbAc318E30258e84Aa880d93b"),
	common.HexToAddress("0xd7aEF28dE2820dEe08A80221a943148316482DfA"),
	common.HexToAddress("0xf84b2C9c80069b2FA10fF7850819bb01bC3ccd1C"),
	common.HexToAddress("0xf7A47902A2F19347552a5b06EAC45e7eA3d5CF9D"),
	common.HexToAddress("0x1BD1136a012312e454CC35F7a80D3b2e4Baf5CFF"),
	common.HexToAddress("0x9b86449941d2f226A2211553f722C76c06d99bbb"),
	common.HexToAddress("0x322fcE1aaF22fBa21b9d0e25CA74ea085866d218"),
	common.HexToAddress("0x19acB57548e036c33Ba3330128E7048d39e175F6"),
	common.HexToAddress("0x173eD01DBE1E6744adf7C1692545664034DE1cf8"),
	common.HexToAddress("0x52dE4c174Df9bF8f70874195090cc6e3aa2A7986"),
	common.HexToAddress("0xF19AEEB8e0cA5A9Cd2597Debaaf6D2baC76A9F0F"),
	common.HexToAddress("0x7A0225fdEe8B1afCf248Ce68C14F087C76F07349"),
	common.HexToAddress("0x28c89F510270A4d536aFEDa16F7240140ACB7C0a"),
	common.HexToAddress("0x66dB00E6DC62c3660D42d69Cdd328dE8cD3A3Db8"),
	common.HexToAddress("0x46F8a89afb8784Cd04A722b4B089f01B35ecE0ca"),
	common.HexToAddress("0xC9A64524049044B403D70eB63655f33aBe2D7Bb2"),
	common.HexToAddress("0xC51bFE53F19120fb926f34fF3a92b6f997FB8f6d"),
	common.HexToAddress("0x53dbe2a81209aAEa1b831aE6Ba5b3c1c9Fe34339"),
	common.HexToAddress("0x50e386588bcd6aa4043482acae5a04a162b0202a"),
	common.HexToAddress("0x9980d59F202fcBd03737d7Cc2223Eb3785A6AD9c"),
	common.HexToAddress("0x512F1423Bf17fe91a18755eFBB3dbAef59e1B7AF"),
	common.HexToAddress("0xA5DC52bFBc753155ACffa37512C974a42bE295fC"),
	common.HexToAddress("0x41E056e572d6662A0128c0e9a3Fc21aBEE9e7967"),
	common.HexToAddress("0x4ec0aD1B1730A7ca472c4afa261C3A3a80e57Fee"),
	common.HexToAddress("0xCB5e12053A1966e964AB1641722acA1e21480C10"),
	common.HexToAddress("0x02f6950fd65299bC526F62948dC1235884024018"),
	common.HexToAddress("0x9980d59F202fcBd03737d7Cc2223Eb3785A6AD9c"),
	common.HexToAddress("0xE712f755e7c5544086383892730C961a245c89B5"),
	common.HexToAddress("0xf8450A9478b49d7fE293f6fbd56bEa899b0eE132"),
	common.HexToAddress("0x3B3Db20319f2d7990f0982da3C1865B87728e15C"),
	common.HexToAddress("0x4dac1ecd9F662deAa46549D1e3A404D0cb9C190C"),
	common.HexToAddress("0x87B8f8Ea433B56995cb6d660291a884fB66396E7"),
	common.HexToAddress("0xd34B9b91F2D467F1e62d978719fbe1c971bd37B3"),
	common.HexToAddress("0xCA70603Ab1f66110841b1F9c77c1335818A5003e"),
	common.HexToAddress("0xa6612755482288F602330f680A2f49D67786B283"),
	common.HexToAddress("0xd700f6379eb255521164233Db7495Bc8c90AE929"),
	common.HexToAddress("0xF9d32237D6af9fc33A8c6E9aBE9f3c2451f25892"),
	common.HexToAddress("0x144ba139E3a1653705B7a38FD4F134eB88C2a5fF"),
	common.HexToAddress("0x14fE4B9B3E18350374e11E27373798DCf98323d0"),
	common.HexToAddress("0x51d26fB6408Cf88C80C2780a5560D7481aD2365C"),
	common.HexToAddress("0x30c90faA26781399cE206a87Ae3cB3BEaA6607E1"),
	common.HexToAddress("0x7E019e2E500D5D84329Eb419a8bE817BE111B8b1"),
	common.HexToAddress("0xEbc5E1739616031Bb9F7307A5301a36e1F0A7e73"),
	common.HexToAddress("0x195d49B3d764d2bd651354C21719D36BD343640B"),
	common.HexToAddress("0x20020B89327C5AeF391bc325F8e27716837DCcE2"),
	common.HexToAddress("0x0E1839A71979490C72Ab79670a02CA93FA642Af5"),
	common.HexToAddress("0xbE8050F2317417b9F9023D39776cC9dF74696131"),
	common.HexToAddress("0xf4D461326Bbb9Aa0AC2F3AECeC7D7Cd362D5F8ea"),
	common.HexToAddress("0xa86efdDd73Fac3165f4234eb512044c92DEdd567"),
	common.HexToAddress("0x998e659A3e28A8957f10086b22683e3020758e7F"),
	common.HexToAddress("0x50e386588BCd6aA4043482aCae5A04A162B0202A"),
	common.HexToAddress("0x445D7A17d2d3344e30C364344C0024e927c2F510"),
	common.HexToAddress("0x24812dd071F259ED388ceFD7394a7e1Af5B2Db9d"),
	common.HexToAddress("0x89c99e57A146E27AB1154e117F171Dd809Cc4FFc"),
	common.HexToAddress("0x8CffdF3980EBCEA1CF6909b000310c2dB7C5E90B"),
	common.HexToAddress("0x2eA418101580e740f0281b1DDCE07223821647b8"),
	common.HexToAddress("0x0833249307d26F4cC3DA7A33FA111E49e0CEf95D"),
	common.HexToAddress("0x9f77D892524BD9167c0A72fd6E58d645e5C14550"),
	common.HexToAddress("0xFCa56e744A2dEa86bC7C540907A22DaD8481f0a4"),
	common.HexToAddress("0xD9e4dD00C947472A6D3dA116E77FaFD0753bB79F"),
	common.HexToAddress("0xEaf94d79cA8F5CCFa06f64ea9E3d0C552A8B208D"),
	common.HexToAddress("0xF0D185E8c2C5688664ae5F71F2f8541eAA257c72"),
	common.HexToAddress("0x9f48a1Dc525ba053dD07124CBC766b44809590CD"),
	common.HexToAddress("0x9f48a1Dc525ba053dD07124CBC766b44809590CD"),
	common.HexToAddress("0x6F02A1D76a7c4B6CfbedDE39FB5D923917723305"),
	common.HexToAddress("0xa9B7F4D601E15DcdA576F78E75045410c8a434E5"),
	common.HexToAddress("0x6316442AAfb9dF3DF006372A825eAC2AEEb7CaF2"),
	common.HexToAddress("0x7939F22785BD4cd6FB05ae2A96BC8cC984Ab5683"),
	common.HexToAddress("0x05d5010Bc05EAf7832043Fbf9A657dbE591579A3"),
	common.HexToAddress("0x98431Aa6718c2ABdC637282E63a183d313CB3d03"), // 484
	common.HexToAddress("0xeBf34eD6a8b719881Dd140a3b10f10D8F9B73B9D"),
	common.HexToAddress("0x2eA4A43d90151F413613D785d11756179509354A"),
	common.HexToAddress("0x46c53f7899b057112D4F8bDb4fe8DF39c8d10Ba9"),
	common.HexToAddress("0xd5414d7d665355F9195cEf20a7725022f01e4867"),
	common.HexToAddress("0x06194691a1f25a5D5Ff19Aa640f18a045EA8356F"),
	common.HexToAddress("0x1B7a7F49FC2d18B0861938Dd9B66934828460E9c"),
	common.HexToAddress("0xa78362d2176438DA6d80ccEA97985B382A05dD09"),
	common.HexToAddress("0xFCa56e744A2dEa86bC7C540907A22DaD8481f0a4"),
	common.HexToAddress("0x600369732F15bbb5d1CF5AB25A3b612Cb5BcC676"),
	common.HexToAddress("0x9b86449941d2f226A2211553f722C76c06d99bbb"),
	common.HexToAddress("0x0784940c6C8196e416052499DCC9fB42B32d85dc"),
	common.HexToAddress("0x56Af4E6De4e70029983966c8ED33b9beFAec3159"),
	common.HexToAddress("0x618D675aF9954f0047Eb2E8AE05BD06F21d3C254"),
	common.HexToAddress("0x721ABC0cE23468573B25338295BC24D66F69b889"),
	common.HexToAddress("0x68a84B6831df19BBF25f3fDf332e0dDF2B745099"),
	common.HexToAddress("0x0cB5C2db4392Ee8f3c2dEE19E768831f132851F2"),
	common.HexToAddress("0xBcc3B346dD3B96074B906Dafaec0c89f51e2D6b2"),
	common.HexToAddress("0xa0837eFF4f3660Aee35C113b4628D6e3f84e0661"),
	common.HexToAddress("0xc36F957313D197522C08bbc5f1cE45B0E212f77e"),
	common.HexToAddress("0x1633A04FcB9F5023152C328FDa13c0a223042870"),
	common.HexToAddress("0xD1aFe1afc32cCb31F195339a16d275097aC5fF2A"),
	common.HexToAddress("0x7A88472Fcc545f0d13586f7Fa8dCE0b8e5F688D9"),
	common.HexToAddress("0x50E4fb14b5D9457766383a36811e7C39D9CE927f"),
	common.HexToAddress("0x2b68887AF525ee1bf8aD309b1bFcB44ae386AFd3"),
	common.HexToAddress("0xb1722BF7F0F62fBa55315B5094610bd18fa14342"),
	common.HexToAddress("0xD250370a9D3f869Ed1427183580ea2Cd1dA7e2Aa"),
	common.HexToAddress("0xFb51A35e332983D0102449FdA4AD51045142c6a9"),
	common.HexToAddress("0x92ef08b5DA2e154fe8c37db5F6feeC62BAA00988"),
	common.HexToAddress("0xc8B6C0638C9343F93D8ab4e9cE5Fa61B7219c1E3"),
	common.HexToAddress("0x46E3836D95431E67092c803294929225aE361D23"),
	common.HexToAddress("0x27D264277dc6684E2Cc2C629FF1166BB2Ee7FC90"),
	common.HexToAddress("0xCc47b1Afcd22F6Fa1EF36B21579eAf73FC400F00"),
	common.HexToAddress("0x0Ca452C4A1da792e6D5C160daDaC59E8068D082e"),
	common.HexToAddress("0x4A85c538D50070cB30B03b704241f5089789452d"),
	common.HexToAddress("0xa01aA1fDE7915636c27627117BB7668861c7B696"),
	common.HexToAddress("0x183c84F6bea52E553a612F1D3555BF7BB762E902"),
	common.HexToAddress("0x0BB93fe7f3de2eb67209F3934587f1680b10C3f6"),
	common.HexToAddress("0x3B87dda21D518CA9Da286e8FE322059Ce9191cd0"),
	common.HexToAddress("0xB4b7201B49a9ae88a24c700C3BF2033F1A14696A"),
	common.HexToAddress("0x6a8437921cf7b1e079d3ab4736cB97816dc7cBB5"),
	common.HexToAddress("0xD61Cfb34bFE2F521D373E4fB9B25f89838db18E2"),
	common.HexToAddress("0x600369732F15bbb5d1CF5AB25A3b612Cb5BcC676"), //526
}

func IsWhitelistedAddress(address string) bool {
	return true
	//addressHex := common.HexToAddress(address)
	//for _, a := range whitelistedAddresses {
	//	if a.Hex() == addressHex.Hex() {
	//		return true
	//	}
	//}

	//return false
}
