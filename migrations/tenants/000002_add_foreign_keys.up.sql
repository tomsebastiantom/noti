ALTER TABLE notifications
ADD CONSTRAINT fk_tenant
FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE notifications
ADD CONSTRAINT fk_user
FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE notifications
ADD CONSTRAINT fk_template
FOREIGN KEY (template_id) REFERENCES templates(id);

ALTER TABLE templates
ADD CONSTRAINT fk_tenant
FOREIGN KEY (tenant_id) REFERENCES tenants(id);

ALTER TABLE users
ADD CONSTRAINT fk_tenant
FOREIGN KEY (tenant_id) REFERENCES tenants(id);
