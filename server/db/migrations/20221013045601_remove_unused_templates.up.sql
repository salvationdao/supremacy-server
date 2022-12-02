DELETE
FROM template_blueprints
WHERE template_id IN (SELECT t.id
                      FROM templates t
                               LEFT OUTER JOIN template_blueprints tb ON t.id = tb.template_id AND "type" = 'MECH_SKIN'
                      WHERE tb IS NULL);

DELETE
FROM templates
WHERE id IN (SELECT t.id
             FROM templates t
                      LEFT OUTER JOIN template_blueprints tb ON t.id = tb.template_id AND "type" = 'MECH_SKIN'
             WHERE tb IS NULL);
