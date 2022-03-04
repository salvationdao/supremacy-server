CREATE TABLE syndicates (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    label TEXT NOT NULL,
    guild_id UUID,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE players (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    syndicate_id UUID NOT NULL REFERENCES syndicates(id),
    public_address TEXT NOT NULL,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE brands (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    syndicate_id UUID NOT NULL REFERENCES syndicates(id),
    label TEXT NOT NULL,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE template_chassis (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    label TEXT NOT NULL,
    slug TEXT NOT NULL,
    shield_recharge_rate INTEGER NOT NULL,
    hp INTEGER NOT NULL,
    brand_id UUID NOT NULL REFERENCES brands(id),
    weapon_hardpoints INTEGER NOT NULL,
    turret_hardpoints INTEGER NOT NULL,
    utility_slots INTEGER NOT NULL,
    speed INTEGER NOT NULL,
    max_hitpoints INTEGER NOT NULL,
    max_shield INTEGER NOT NULL,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE templates (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    template_chassis_id UUID NOT NULL REFERENCES template_chassis(id),
    label TEXT NOT NULL,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE template_weapons (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    label TEXT NOT NULL,
    slug TEXT NOT NULL,
    damage INTEGER NOT NULL,
    weapon_type TEXT NOT NULL CHECK (weapon_type IN ('SHOULDER', 'ARM')),
    brand_id UUID NOT NULL REFERENCES brands(id),

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE template_modules (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL,
    label TEXT NOT NULL,
    hp_modifier INTEGER NOT NULL,
    shield_modifier INTEGER NOT NULL,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE templates_template_weapons (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    template_weapon_id UUID NOT NULL REFERENCES template_weapons(id),
    template_id UUID NOT NULL REFERENCES templates(id),

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE templates_template_modules (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE mechs (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES players(id),
    label TEXT NOT NULL,
    health_remaining INTEGER NOT NULL,
    skin TEXT NOT NULL,
    model TEXT NOT NULL,
    brand_id UUID NOT NULL REFERENCES brands(id),
    slug TEXT NOT NULL,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE weapons (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    label TEXT NOT NULL,
    slug TEXT NOT NULL,
    damage INTEGER NOT NULL,
    weapon_type TEXT NOT NULL CHECK (weapon_type IN ('SHOULDER', 'ARM')),
    brand_id UUID NOT NULL REFERENCES brands(id),

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chassis (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    label TEXT NOT NULL,
    slug TEXT NOT NULL,
    shield_recharge_rate INTEGER NOT NULL,
    hp INTEGER NOT NULL,
    brand_id UUID NOT NULL REFERENCES brands(id),
    weapon_hardpoints INTEGER NOT NULL,
    turret_hardpoints INTEGER NOT NULL,
    utility_slots INTEGER NOT NULL,
    speed INTEGER NOT NULL,
    max_hitpoints INTEGER NOT NULL,
    max_shield INTEGER NOT NULL,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE modules (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL,
    label TEXT NOT NULL,
    hp_modifier INTEGER NOT NULL,
    shield_modifier INTEGER NOT NULL,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE mechs_weapons (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    mech_id UUID NOT NULL REFERENCES mechs(id),
    weapon_id UUID NOT NULL REFERENCES weapons(id),
    slot_number INTEGER NOT NULL,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE mechs_modules (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    mech_id UUID NOT NULL REFERENCES mechs(id),
    module_id UUID NOT NULL REFERENCES modules(id),
    slot_number INTEGER NOT NULL,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);