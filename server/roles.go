package server

import (
	"time"

	"github.com/gofrs/uuid"
)

var (
	UserRoleBroadcasterID = RoleID(uuid.Must(uuid.FromString("cca82653-c071-4171-92da-05b0808542e7")))
	UserRoleViewerID      = RoleID(uuid.Must(uuid.FromString("1fb981b2-7489-4061-a379-1430ec4f7a63")))
	UserRoleSuperAdminID  = RoleID(uuid.Must(uuid.FromString("85837f44-988c-4d1d-a292-e376b87015cd")))
	UserRoleAdmin         = RoleID(uuid.Must(uuid.FromString("7e8f0c1d-f36c-437c-bee2-c14fedb4df93")))
	UserRolePlayer        = RoleID(uuid.Must(uuid.FromString("8dd55355-fc22-4d1d-a825-b973bb075259")))
)

// Role is an object representing the database table.
type Role struct {
	ID          RoleID     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Permissions []string   `json:"permissions" db:"permissions"`
	Tier        int        `json:"tier" db:"tier"`
	Reserved    bool       `json:"reserved" db:"reserved"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at" db:"deleted_at"`
}

// Perm permission enum
type Perm string

// Permission enums
const (
	PermRoleList      Perm = "RoleList"
	PermRoleCreate    Perm = "RoleCreate"
	PermRoleRead      Perm = "RoleRead"
	PermRoleUpdate    Perm = "RoleUpdate"
	PermRoleArchive   Perm = "RoleArchive"
	PermRoleUnarchive Perm = "RoleUnarchive"

	PermUserList            Perm = "UserList"
	PermUserCreate          Perm = "UserCreate"
	PermUserRead            Perm = "UserRead"
	PermUserUpdate          Perm = "UserUpdate"
	PermUserArchive         Perm = "UserArchive"
	PermUserUnarchive       Perm = "UserUnarchive"
	PermUserForceDisconnect Perm = "UserForceDisconnect"

	PermOrganisationList      Perm = "OrganisationList"
	PermOrganisationCreate    Perm = "OrganisationCreate"
	PermOrganisationRead      Perm = "OrganisationRead"
	PermOrganisationUpdate    Perm = "OrganisationUpdate"
	PermOrganisationArchive   Perm = "OrganisationArchive"
	PermOrganisationUnarchive Perm = "OrganisationUnarchive"

	PermProductList      Perm = "ProductList"
	PermProductCreate    Perm = "ProductCreate"
	PermProductRead      Perm = "ProductRead"
	PermProductUpdate    Perm = "ProductUpdate"
	PermProductArchive   Perm = "ProductArchive"
	PermProductUnarchive Perm = "ProductUnarchive"

	PermAdminPortal      Perm = "AdminPortal"
	PermImpersonateUser  Perm = "ImpersonateUser"
	PermUserActivityList Perm = "UserActivityList"
)

// AllPerm contains all permissions
var AllPerm = []Perm{
	PermRoleList,
	PermRoleCreate,
	PermRoleRead,
	PermRoleUpdate,
	PermRoleArchive,
	PermRoleUnarchive,

	PermUserList,
	PermUserCreate,
	PermUserRead,
	PermUserUpdate,
	PermUserArchive,
	PermUserUnarchive,
	PermUserForceDisconnect,

	PermOrganisationList,
	PermOrganisationCreate,
	PermOrganisationRead,
	PermOrganisationUpdate,
	PermOrganisationArchive,
	PermOrganisationUnarchive,

	PermProductList,
	PermProductCreate,
	PermProductRead,
	PermProductUpdate,
	PermProductArchive,
	PermProductUnarchive,

	PermAdminPortal,
	PermImpersonateUser,
	PermUserActivityList,
}

func (e Perm) String() string {
	return string(e)
}
