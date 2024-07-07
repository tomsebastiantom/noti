ALTER TABLE notifications
DROP CONSTRAINT fk_tenant;

ALTER TABLE notifications
DROP CONSTRAINT fk_user;

ALTER TABLE notifications
DROP CONSTRAINT fk_template;

ALTER TABLE templates
DROP CONSTRAINT fk_tenant;

ALTER TABLE users
DROP CONSTRAINT fk_tenant;
