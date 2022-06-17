WITH ms AS (SELECT id, image_url, avatar_url, large_image_url, animation_url, card_animation_url FROM mech_skin)
UPDATE collection_items ci
SET image_url          = ms.image_url,
    avatar_url         = ms.avatar_url,
    large_image_url    = ms.large_image_url,
    animation_url      = ms.animation_url,
    card_animation_url = ms.card_animation_url
FROM ms
WHERE ci.item_id = ms.id;