-- Seed ClickSafe testing database

-- 1. Insert Department
INSERT INTO departments (id, label, attributes)
VALUES ('e5270921-2a1f-4bb0-8a19-482470eb0001', 'ИТ-департамент', '{}'::jsonb)
ON CONFLICT (id) DO UPDATE SET label = EXCLUDED.label, attributes = EXCLUDED.attributes;

-- 2. Insert Employees
INSERT INTO employees (id, department_id, first_name, last_name, email, phone_number, attributes)
VALUES
('e5270921-2a1f-4bb0-8a19-482470eb0002', 'e5270921-2a1f-4bb0-8a19-482470eb0001', 'Иван', 'Иванов', 'ivanov@company.com', '+79998887766', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0003', 'e5270921-2a1f-4bb0-8a19-482470eb0001', 'Мария', 'Петрова', 'petrova@company.com', '+79997776655', '{}'::jsonb)
ON CONFLICT (id) DO UPDATE SET
    department_id = EXCLUDED.department_id,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    email = EXCLUDED.email,
    phone_number = EXCLUDED.phone_number,
    attributes = EXCLUDED.attributes;

-- 3. Insert Attachments
INSERT INTO attachments (id, label, type, content_path, required_vars, uploaded_at, is_public)
VALUES
('e5270921-2a1f-4bb0-8a19-482470eb0004', 'Microsoft Sign-In Form', '.html', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0004.html', ARRAY['Employee.Email.Address', 'Employee.FirstName', 'Employee.LastName', 'Target.Link'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0005', 'Security Awareness Warning', '.html', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0005.html', ARRAY['Employee.FirstName', 'Employee.LastName', 'Department.Label'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0008', 'MAX: Смена пароля (со ссылкой) - шаблон', '.txt', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0008.txt', ARRAY['Employee.FirstName', 'Target.Link'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0009', 'MAX: Сверка СНИЛС (без ссылки) - шаблон', '.txt', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0009.txt', ARRAY['Employee.FirstName'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0012', 'MAX: Обучающее сообщение', '.txt', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0012.txt', ARRAY['Employee.FirstName'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0013', 'MAX: Обновление КриптоПро (со ссылкой) - шаблон', '.txt', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0013.txt', ARRAY['Employee.FirstName', 'Employee.LastName', 'Target.Link'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0014', 'MAX: Истечение пароля (со ссылкой) - шаблон', '.txt', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0014.txt', ARRAY['Employee.FirstName', 'Target.Link'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0015', 'EMAIL: Установка обновления безопасности ИСиР - шаблон', '.html', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0015.html', ARRAY['Employee.FirstName', 'Employee.LastName', 'Target.Link'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0016', 'LANDING: Авторизация ИСиР (ДИТ)', '.html', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0016.html', ARRAY['Target.Link'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0022', 'MAX: График отпусков (без ссылки) - шаблон', '.txt', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0022.txt', ARRAY['Employee.FirstName'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0023', 'MAX: Обучение (График отпусков)', '.txt', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0023.txt', ARRAY['Employee.FirstName'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0024', 'EMAIL: График отпусков - шаблон', '.html', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0024.html', ARRAY['Employee.FirstName', 'Employee.LastName', 'Target.Link'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0025', 'LANDING: HR Портал (Отпуска) - шаблон', '.html', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0025.html', ARRAY['Target.Link'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0026', 'EDUCATION: Обучение (HR Фишинг)', '.html', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0026.html', ARRAY['Employee.FirstName'], NOW(), FALSE)
ON CONFLICT (id) DO UPDATE SET
    label = EXCLUDED.label,
    type = EXCLUDED.type,
    content_path = EXCLUDED.content_path,
    required_vars = EXCLUDED.required_vars,
    uploaded_at = EXCLUDED.uploaded_at,
    is_public = EXCLUDED.is_public;

-- 4. Insert Landings
INSERT INTO landings (id, label, html_body_id)
VALUES
('e5270921-2a1f-4bb0-8a19-482470eb0006', 'Microsoft Login Page', 'e5270921-2a1f-4bb0-8a19-482470eb0004'),
('e5270921-2a1f-4bb0-8a19-482470eb0007', 'Security Warning Education', 'e5270921-2a1f-4bb0-8a19-482470eb0005'),
('e5270921-2a1f-4bb0-8a19-482470eb0017', 'Авторизация ИСиР (ДИТ)', 'e5270921-2a1f-4bb0-8a19-482470eb0016'),
('e5270921-2a1f-4bb0-8a19-482470eb0027', 'HR Портал (График отпусков)', 'e5270921-2a1f-4bb0-8a19-482470eb0025'),
('e5270921-2a1f-4bb0-8a19-482470eb0028', 'Обучение (График отпусков)', 'e5270921-2a1f-4bb0-8a19-482470eb0026')
ON CONFLICT (id) DO UPDATE SET
    label = EXCLUDED.label,
    html_body_id = EXCLUDED.html_body_id;

-- 5. Insert MAX Messages
INSERT INTO messages (id, label, from_email, from_name, subject, type, max_account_id, text_body_id, html_body_id)
VALUES
('e5270921-2a1f-4bb0-8a19-482470eb0010', 'MAX: Смена пароля (со ссылкой)', 'security@company.com', 'Система безопасности', 'Срочная смена пароля', 'MAX', (SELECT id FROM max_accounts LIMIT 1), 'e5270921-2a1f-4bb0-8a19-482470eb0008', NULL),
('e5270921-2a1f-4bb0-8a19-482470eb0011', 'MAX: Сверка СНИЛС (без ссылки)', 'hr@company.com', 'Отдел кадров', 'Сверка личных данных', 'MAX', (SELECT id FROM max_accounts LIMIT 1), 'e5270921-2a1f-4bb0-8a19-482470eb0009', NULL),
('e5270921-2a1f-4bb0-8a19-482470eb0018', 'MAX: Обновление КриптоПро (со ссылкой)', 'support-dit@mos-transport.ru', 'ДИТ Москвы', 'Обновление КриптоПро NGate', 'MAX', (SELECT id FROM max_accounts LIMIT 1), 'e5270921-2a1f-4bb0-8a19-482470eb0013', NULL),
('e5270921-2a1f-4bb0-8a19-482470eb0019', 'MAX: Истечение пароля (со ссылкой)', 'security@mosmetro-internal.ru', 'Подразделение ИБ', 'Уведомление об истечении пароля', 'MAX', (SELECT id FROM max_accounts LIMIT 1), 'e5270921-2a1f-4bb0-8a19-482470eb0014', NULL),
('e5270921-2a1f-4bb0-8a19-482470eb0020', 'EMAIL: Установка обновления безопасности ИСиР', 'dit@mos.ru', 'Департамент информационных технологий', 'Срочно: Установка обновления безопасности (TLS и КриптоПро)', 'EMAIL', NULL, NULL, 'e5270921-2a1f-4bb0-8a19-482470eb0015'),
('e5270921-2a1f-4bb0-8a19-482470eb0021', 'MAX: Сверка графика отпусков (без ссылки)', 'hr@company.com', 'Отдел кадров', 'График отпусков', 'MAX', (SELECT id FROM max_accounts LIMIT 1), 'e5270921-2a1f-4bb0-8a19-482470eb0022', NULL),
('e5270921-2a1f-4bb0-8a19-482470eb0029', 'EMAIL: Электронное подписание графика отпусков', 'hr@company.com', 'Отдел кадров', 'Ваш график отпусков ожидает подписания', 'EMAIL', NULL, NULL, 'e5270921-2a1f-4bb0-8a19-482470eb0024')
ON CONFLICT (id) DO UPDATE SET
    label = EXCLUDED.label,
    from_email = EXCLUDED.from_email,
    from_name = EXCLUDED.from_name,
    subject = EXCLUDED.subject,
    type = EXCLUDED.type,
    max_account_id = EXCLUDED.max_account_id,
    text_body_id = EXCLUDED.text_body_id,
    html_body_id = EXCLUDED.html_body_id;
