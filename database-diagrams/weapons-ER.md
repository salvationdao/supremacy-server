classDiagram
direction BT
class blueprint_weapon_skin {
   text label
   text tier
   timestamp with time zone created_at
   text collection
   numeric(8) stat_modifier
   uuid id
}
class blueprint_weapons {
   text label
   text slug
   integer damage
   timestamp with time zone deleted_at
   timestamp with time zone updated_at
   timestamp with time zone created_at
   uuid game_client_weapon_id
   weapon_type weapon_type
   collection collection
   damage_type default_damage_type
   integer damage_falloff
   integer damage_falloff_rate
   integer radius
   integer radius_damage_falloff
   numeric spread
   numeric rate_of_fire
   numeric projectile_speed
   integer max_ammo
   boolean is_melee
   text tier
   numeric energy_cost
   uuid weapon_model_id
   uuid id
}
class brands {
   uuid faction_id
   text label
   timestamp with time zone deleted_at
   timestamp with time zone updated_at
   timestamp with time zone created_at
   uuid id
}
class factions {
   text vote_price
   text contract_reward
   text label
   uuid guild_id
   timestamp with time zone deleted_at
   timestamp with time zone updated_at
   timestamp with time zone created_at
   text primary_color
   text secondary_color
   text background_color
   text logo_url
   text background_url
   text description
   text wallpaper_url
   uuid id
}
class weapon_model_skin_compatibilities {
   text image_url
   text card_animation_url
   text avatar_url
   text large_image_url
   text background_color
   text animation_url
   text youtube_url
   timestamp with time zone deleted_at
   timestamp with time zone updated_at
   timestamp with time zone created_at
   uuid blueprint_weapon_skin_id
   uuid weapon_model_id
}
class weapon_models {
   uuid brand_id
   text label
   weapon_type weapon_type
   uuid default_skin_id
   timestamp with time zone deleted_at
   timestamp with time zone updated_at
   timestamp with time zone created_at
   integer repair_blocks
   uuid id
}
class weapon_skin {
   uuid blueprint_id
   uuid equipped_on
   timestamp with time zone created_at
   uuid id
}
class weapons {
   text slug
   integer damage
   timestamp with time zone deleted_at
   timestamp with time zone updated_at
   timestamp with time zone created_at
   uuid blueprint_id
   uuid equipped_on
   damage_type default_damage_type
   bigint genesis_token_id
   bigint limited_release_token_id
   integer damage_falloff
   integer damage_falloff_rate
   integer radius
   integer radius_damage_falloff
   numeric spread
   numeric rate_of_fire
   numeric projectile_speed
   numeric energy_cost
   boolean is_melee
   integer max_ammo
   boolean locked_to_mech
   uuid equipped_weapon_skin_id
   uuid id
}

blueprint_weapons  -->  weapon_models : weapon_model_id:id
brands  -->  factions : faction_id:id
weapon_model_skin_compatibilities  -->  blueprint_weapon_skin : blueprint_weapon_skin_id:id
weapon_model_skin_compatibilities  -->  weapon_models : weapon_model_id:id
weapon_models  -->  blueprint_weapon_skin : default_skin_id:id
weapon_models  -->  brands : brand_id:id
weapon_skin  -->  blueprint_weapon_skin : blueprint_id:id
weapon_skin  -->  weapons : equipped_on:id
weapons  -->  blueprint_weapons : blueprint_id:id
weapons  -->  weapon_skin : equipped_weapon_skin_id:id
