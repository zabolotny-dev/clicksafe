-- Seed ClickSafe testing database

-- 0. Update organization with Metro branding
INSERT INTO attachments (id, label, type, content_path, required_vars, uploaded_at, is_public)
VALUES ('e5270921-2a1f-4bb0-8a19-482470eb0034', 'Логотип Московского метрополитена', '.png', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0034.png', ARRAY[]::text[], NOW(), TRUE)
ON CONFLICT (id) DO UPDATE SET
    label = EXCLUDED.label,
    type = EXCLUDED.type,
    content_path = EXCLUDED.content_path,
    is_public = EXCLUDED.is_public;

UPDATE organizations
SET label = 'ГУП «Московский метрополитен»',
    attachment_id = 'e5270921-2a1f-4bb0-8a19-482470eb0034'
WHERE id = '00000000-0000-0000-0000-000000000001';

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
('e5270921-2a1f-4bb0-8a19-482470eb0025', 'LANDING: HR Портал (Отпуска) - шаблон', '.html', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0025.html', ARRAY['Target.Link', 'Organization.Logo'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0026', 'EDUCATION: Обучение (HR Фишинг)', '.html', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0026.html', ARRAY['Employee.FirstName'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0030', 'LANDING: Портал самообслуживания (Смена пароля)', '.html', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0030.html', ARRAY['Target.Link', 'Organization.Logo'], NOW(), FALSE),
('e5270921-2a1f-4bb0-8a19-482470eb0031', 'EDUCATION: Обучение (Безопасность паролей)', '.html', '/attachment/e5270921-2a1f-4bb0-8a19-482470eb0031.html', ARRAY['Employee.FirstName'], NOW(), FALSE)
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
('e5270921-2a1f-4bb0-8a19-482470eb0028', 'Обучение (График отпусков)', 'e5270921-2a1f-4bb0-8a19-482470eb0026'),
('e5270921-2a1f-4bb0-8a19-482470eb0032', 'Портал самообслуживания (Смена пароля)', 'e5270921-2a1f-4bb0-8a19-482470eb0030'),
('e5270921-2a1f-4bb0-8a19-482470eb0033', 'Обучение (Безопасность паролей)', 'e5270921-2a1f-4bb0-8a19-482470eb0031')
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

-- 6. Additional Departments
INSERT INTO departments (id, label, attributes)
VALUES
('e5270921-2a1f-4bb0-8a19-482470eb0040', 'HR-отдел', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0041', 'Бухгалтерия', '{}'::jsonb)
ON CONFLICT (id) DO UPDATE SET label = EXCLUDED.label, attributes = EXCLUDED.attributes;

-- 7. Additional Employees
INSERT INTO employees (id, department_id, first_name, last_name, email, phone_number, attributes)
VALUES
('e5270921-2a1f-4bb0-8a19-482470eb0042', 'e5270921-2a1f-4bb0-8a19-482470eb0001', 'Алексей',   'Смирнов',   'smirnov@company.com',   '+79996665544', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0043', 'e5270921-2a1f-4bb0-8a19-482470eb0001', 'Елена',     'Козлова',   'kozlova@company.com',   '+79995554433', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0044', 'e5270921-2a1f-4bb0-8a19-482470eb0001', 'Дмитрий',   'Новиков',   'novikov@company.com',   '+79994443322', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0045', 'e5270921-2a1f-4bb0-8a19-482470eb0040', 'Ольга',     'Морозова',  'morozova@company.com',  '+79993332211', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0046', 'e5270921-2a1f-4bb0-8a19-482470eb0040', 'Сергей',    'Волков',    'volkov@company.com',    '+79992221100', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0047', 'e5270921-2a1f-4bb0-8a19-482470eb0040', 'Наталья',   'Лебедева',  'lebedeva@company.com',  '+79991110099', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0048', 'e5270921-2a1f-4bb0-8a19-482470eb0040', 'Андрей',    'Соколов',   'sokolov@company.com',   '+79990009988', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0049', 'e5270921-2a1f-4bb0-8a19-482470eb0041', 'Татьяна',   'Попова',    'popova@company.com',    '+79989998877', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0050', 'e5270921-2a1f-4bb0-8a19-482470eb0041', 'Михаил',    'Захаров',   'zakharov@company.com',  '+79988887766', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0051', 'e5270921-2a1f-4bb0-8a19-482470eb0041', 'Виктория',  'Кузнецова', 'kuznetsova@company.com', '+79987776655', '{}'::jsonb),
('e5270921-2a1f-4bb0-8a19-482470eb0052', 'e5270921-2a1f-4bb0-8a19-482470eb0040', 'Наталья',   'Петренко',  'petrenko@company.com',  '+79163115104', '{}'::jsonb)
ON CONFLICT (id) DO UPDATE SET
    department_id = EXCLUDED.department_id,
    first_name    = EXCLUDED.first_name,
    last_name     = EXCLUDED.last_name,
    email         = EXCLUDED.email,
    phone_number  = EXCLUDED.phone_number,
    attributes    = EXCLUDED.attributes;

-- 8. Campaigns
INSERT INTO campaigns (id, message_id, landing_id, education_id, max_education_text_id, domain, label, type, status, date_from, date_to, attributes)
VALUES
(
    'e5270921-2a1f-4bb0-8a19-482470eb0060',
    'e5270921-2a1f-4bb0-8a19-482470eb0020',
    'e5270921-2a1f-4bb0-8a19-482470eb0017',
    'e5270921-2a1f-4bb0-8a19-482470eb0007',
    NULL,
    'https://dit-mos-update.ru',
    'Фишинг: Обновление ИСиР',
    'EMAIL', 'COMPLETED',
    '2026-05-12 06:00:00+00', '2026-05-18 15:00:00+00',
    '{}'::jsonb
),
(
    'e5270921-2a1f-4bb0-8a19-482470eb0062',
    'e5270921-2a1f-4bb0-8a19-482470eb0029',
    'e5270921-2a1f-4bb0-8a19-482470eb0032',
    'e5270921-2a1f-4bb0-8a19-482470eb0033',
    NULL,
    'https://vacation-sign-portal.ru',
    'Фишинг: Подписание графика отпусков',
    'EMAIL', 'COMPLETED',
    '2026-06-02 06:00:00+00', '2026-06-07 15:00:00+00',
    '{}'::jsonb
)
ON CONFLICT (id) DO UPDATE SET
    message_id           = EXCLUDED.message_id,
    landing_id           = EXCLUDED.landing_id,
    education_id         = EXCLUDED.education_id,
    max_education_text_id = EXCLUDED.max_education_text_id,
    domain               = EXCLUDED.domain,
    label                = EXCLUDED.label,
    type                 = EXCLUDED.type,
    status               = EXCLUDED.status,
    date_from            = EXCLUDED.date_from,
    date_to              = EXCLUDED.date_to,
    attributes           = EXCLUDED.attributes;

-- 9. Targets
-- Campaign 1 (EMAIL: Обновление ИСиР, 12-18 May)
INSERT INTO targets (id, token, employee_id, campaign_id, status, scheduled_at, created_at)
VALUES
('e5270921-2a1f-4bb0-8a19-482470eb0070','c1t01ivanov',   'e5270921-2a1f-4bb0-8a19-482470eb0002','e5270921-2a1f-4bb0-8a19-482470eb0060','SUBMITTED','2026-05-12 09:00:00+00','2026-05-11 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0071','c1t02petrova',  'e5270921-2a1f-4bb0-8a19-482470eb0003','e5270921-2a1f-4bb0-8a19-482470eb0060','CLICKED',  '2026-05-12 09:00:00+00','2026-05-11 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0072','c1t03smirnov',  'e5270921-2a1f-4bb0-8a19-482470eb0042','e5270921-2a1f-4bb0-8a19-482470eb0060','SUBMITTED','2026-05-12 09:00:00+00','2026-05-11 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0073','c1t04kozlova',  'e5270921-2a1f-4bb0-8a19-482470eb0043','e5270921-2a1f-4bb0-8a19-482470eb0060','OPENED',   '2026-05-12 09:30:00+00','2026-05-11 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0074','c1t05novikov',  'e5270921-2a1f-4bb0-8a19-482470eb0044','e5270921-2a1f-4bb0-8a19-482470eb0060','SENT',     '2026-05-12 10:00:00+00','2026-05-11 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0075','c1t06morozova', 'e5270921-2a1f-4bb0-8a19-482470eb0045','e5270921-2a1f-4bb0-8a19-482470eb0060','CLICKED',  '2026-05-13 09:00:00+00','2026-05-11 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0076','c1t07volkov',   'e5270921-2a1f-4bb0-8a19-482470eb0046','e5270921-2a1f-4bb0-8a19-482470eb0060','SENT',     '2026-05-13 09:00:00+00','2026-05-11 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0077','c1t08lebedeva', 'e5270921-2a1f-4bb0-8a19-482470eb0047','e5270921-2a1f-4bb0-8a19-482470eb0060','OPENED',   '2026-05-13 10:00:00+00','2026-05-11 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0078','c1t09sokolov',  'e5270921-2a1f-4bb0-8a19-482470eb0048','e5270921-2a1f-4bb0-8a19-482470eb0060','SENT',     '2026-05-14 09:00:00+00','2026-05-11 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0079','c1t10popova',   'e5270921-2a1f-4bb0-8a19-482470eb0049','e5270921-2a1f-4bb0-8a19-482470eb0060','SUBMITTED','2026-05-14 09:00:00+00','2026-05-11 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0080','c1t11zakharov', 'e5270921-2a1f-4bb0-8a19-482470eb0050','e5270921-2a1f-4bb0-8a19-482470eb0060','SENT',     '2026-05-15 09:00:00+00','2026-05-11 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0081','c1t12kuznets',  'e5270921-2a1f-4bb0-8a19-482470eb0051','e5270921-2a1f-4bb0-8a19-482470eb0060','OPENED',   '2026-05-15 09:00:00+00','2026-05-11 15:00:00+00')
ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, scheduled_at = EXCLUDED.scheduled_at;

-- Campaign 2 (EMAIL: Подписание графика отпусков, 2-7 Jun)
INSERT INTO targets (id, token, employee_id, campaign_id, status, scheduled_at, created_at)
VALUES
('e5270921-2a1f-4bb0-8a19-482470eb0094','c3t01ivanov',   'e5270921-2a1f-4bb0-8a19-482470eb0002','e5270921-2a1f-4bb0-8a19-482470eb0062','SENT',     '2026-06-02 09:00:00+00','2026-06-01 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0095','c3t02petrova',  'e5270921-2a1f-4bb0-8a19-482470eb0003','e5270921-2a1f-4bb0-8a19-482470eb0062','CLICKED',  '2026-06-02 09:00:00+00','2026-06-01 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0096','c3t03smirnov',  'e5270921-2a1f-4bb0-8a19-482470eb0042','e5270921-2a1f-4bb0-8a19-482470eb0062','SUBMITTED','2026-06-02 09:30:00+00','2026-06-01 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0097','c3t04kozlova',  'e5270921-2a1f-4bb0-8a19-482470eb0043','e5270921-2a1f-4bb0-8a19-482470eb0062','OPENED',   '2026-06-02 10:00:00+00','2026-06-01 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0098','c3t05novikov',  'e5270921-2a1f-4bb0-8a19-482470eb0044','e5270921-2a1f-4bb0-8a19-482470eb0062','SENT',     '2026-06-03 09:00:00+00','2026-06-01 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0099','c3t06morozova', 'e5270921-2a1f-4bb0-8a19-482470eb0045','e5270921-2a1f-4bb0-8a19-482470eb0062','SENT',     '2026-06-03 09:00:00+00','2026-06-01 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0100','c3t07volkov',   'e5270921-2a1f-4bb0-8a19-482470eb0046','e5270921-2a1f-4bb0-8a19-482470eb0062','OPENED',   '2026-06-03 10:00:00+00','2026-06-01 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0101','c3t08lebedeva', 'e5270921-2a1f-4bb0-8a19-482470eb0047','e5270921-2a1f-4bb0-8a19-482470eb0062','CLICKED',  '2026-06-04 09:00:00+00','2026-06-01 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0102','c3t09sokolov',  'e5270921-2a1f-4bb0-8a19-482470eb0048','e5270921-2a1f-4bb0-8a19-482470eb0062','SENT',     '2026-06-04 09:00:00+00','2026-06-01 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0103','c3t10popova',   'e5270921-2a1f-4bb0-8a19-482470eb0049','e5270921-2a1f-4bb0-8a19-482470eb0062','SENT',     '2026-06-05 09:00:00+00','2026-06-01 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0104','c3t11zakharov', 'e5270921-2a1f-4bb0-8a19-482470eb0050','e5270921-2a1f-4bb0-8a19-482470eb0062','OPENED',   '2026-06-05 09:00:00+00','2026-06-01 15:00:00+00'),
('e5270921-2a1f-4bb0-8a19-482470eb0105','c3t12kuznets',  'e5270921-2a1f-4bb0-8a19-482470eb0051','e5270921-2a1f-4bb0-8a19-482470eb0062','SENT',     '2026-06-06 09:00:00+00','2026-06-01 15:00:00+00')
ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, scheduled_at = EXCLUDED.scheduled_at;

-- 10. Events
-- Campaign 1: EMAIL_OPENED / LINK_OPENED / DATA_SENT
INSERT INTO events (id, campaign_id, employee_id, type, ip_address, user_agent, referer, occurred_at)
VALUES
-- Иванов: SUBMITTED
('e5270921-2a1f-4bb0-8a19-482470eb0106','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0002','EMAIL_OPENED','10.10.1.15',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', NULL,                          '2026-05-12 09:47:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0107','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0002','LINK_OPENED', '10.10.1.15',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', 'https://dit-mos-update.ru/', '2026-05-12 09:49:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0108','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0002','DATA_SENT',   '10.10.1.15',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', 'https://dit-mos-update.ru/', '2026-05-12 09:52:00'),
-- Петрова: CLICKED
('e5270921-2a1f-4bb0-8a19-482470eb0109','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0003','EMAIL_OPENED','10.10.1.22',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', NULL,                          '2026-05-12 10:15:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0110','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0003','LINK_OPENED', '10.10.1.22',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', 'https://dit-mos-update.ru/', '2026-05-12 10:17:00'),
-- Смирнов: SUBMITTED
('e5270921-2a1f-4bb0-8a19-482470eb0111','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0042','EMAIL_OPENED','10.10.2.5',   'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/125.0 Safari/537.36', NULL,                          '2026-05-12 11:03:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0112','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0042','LINK_OPENED', '10.10.2.5',   'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/125.0 Safari/537.36', 'https://dit-mos-update.ru/', '2026-05-12 11:05:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0113','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0042','DATA_SENT',   '10.10.2.5',   'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/125.0 Safari/537.36', 'https://dit-mos-update.ru/', '2026-05-12 11:08:00'),
-- Козлова: OPENED
('e5270921-2a1f-4bb0-8a19-482470eb0114','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0043','EMAIL_OPENED','10.10.2.31',  'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', NULL,                         '2026-05-13 08:22:00'),
-- Морозова: CLICKED
('e5270921-2a1f-4bb0-8a19-482470eb0115','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0045','EMAIL_OPENED','192.168.3.44','Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', NULL,                          '2026-05-13 09:34:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0116','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0045','LINK_OPENED', '192.168.3.44','Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', 'https://dit-mos-update.ru/', '2026-05-13 09:37:00'),
-- Лебедева: OPENED
('e5270921-2a1f-4bb0-8a19-482470eb0117','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0047','EMAIL_OPENED','10.10.3.18',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', NULL,                          '2026-05-14 10:05:00'),
-- Попова: SUBMITTED
('e5270921-2a1f-4bb0-8a19-482470eb0118','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0049','EMAIL_OPENED','10.10.4.7',   'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/123.0 Safari/537.36', NULL,                          '2026-05-14 13:20:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0119','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0049','LINK_OPENED', '10.10.4.7',   'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/123.0 Safari/537.36', 'https://dit-mos-update.ru/', '2026-05-14 13:23:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0120','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0049','DATA_SENT',   '10.10.4.7',   'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/123.0 Safari/537.36', 'https://dit-mos-update.ru/', '2026-05-14 13:26:00'),
-- Кузнецова: OPENED
('e5270921-2a1f-4bb0-8a19-482470eb0121','e5270921-2a1f-4bb0-8a19-482470eb0060','e5270921-2a1f-4bb0-8a19-482470eb0051','EMAIL_OPENED','10.10.4.55',  'Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15', NULL,          '2026-05-15 07:48:00'),

-- Campaign 2: EMAIL_OPENED / LINK_OPENED / DATA_SENT
-- Петрова: CLICKED
('e5270921-2a1f-4bb0-8a19-482470eb0148','e5270921-2a1f-4bb0-8a19-482470eb0062','e5270921-2a1f-4bb0-8a19-482470eb0003','EMAIL_OPENED','10.10.1.22',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/125.0 Safari/537.36', NULL,                               '2026-06-02 10:04:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0149','e5270921-2a1f-4bb0-8a19-482470eb0062','e5270921-2a1f-4bb0-8a19-482470eb0003','LINK_OPENED', '10.10.1.22',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/125.0 Safari/537.36', 'https://vacation-sign-portal.ru/', '2026-06-02 10:07:00'),
-- Смирнов: SUBMITTED
('e5270921-2a1f-4bb0-8a19-482470eb0150','e5270921-2a1f-4bb0-8a19-482470eb0062','e5270921-2a1f-4bb0-8a19-482470eb0042','EMAIL_OPENED','10.10.2.5',   'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/125.0 Safari/537.36', NULL,                               '2026-06-02 11:30:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0151','e5270921-2a1f-4bb0-8a19-482470eb0062','e5270921-2a1f-4bb0-8a19-482470eb0042','LINK_OPENED', '10.10.2.5',   'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/125.0 Safari/537.36', 'https://vacation-sign-portal.ru/', '2026-06-02 11:33:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0152','e5270921-2a1f-4bb0-8a19-482470eb0062','e5270921-2a1f-4bb0-8a19-482470eb0042','DATA_SENT',   '10.10.2.5',   'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/125.0 Safari/537.36', 'https://vacation-sign-portal.ru/', '2026-06-02 11:36:00'),
-- Козлова: OPENED
('e5270921-2a1f-4bb0-8a19-482470eb0153','e5270921-2a1f-4bb0-8a19-482470eb0062','e5270921-2a1f-4bb0-8a19-482470eb0043','EMAIL_OPENED','10.10.2.31',  'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', NULL,                            '2026-06-03 09:15:00'),
-- Волков: OPENED
('e5270921-2a1f-4bb0-8a19-482470eb0154','e5270921-2a1f-4bb0-8a19-482470eb0062','e5270921-2a1f-4bb0-8a19-482470eb0046','EMAIL_OPENED','192.168.3.44','Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', NULL,                               '2026-06-03 14:20:00'),
-- Лебедева: CLICKED
('e5270921-2a1f-4bb0-8a19-482470eb0155','e5270921-2a1f-4bb0-8a19-482470eb0062','e5270921-2a1f-4bb0-8a19-482470eb0047','EMAIL_OPENED','10.10.3.18',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/125.0 Safari/537.36', NULL,                               '2026-06-04 08:55:00'),
('e5270921-2a1f-4bb0-8a19-482470eb0156','e5270921-2a1f-4bb0-8a19-482470eb0062','e5270921-2a1f-4bb0-8a19-482470eb0047','LINK_OPENED', '10.10.3.18',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/125.0 Safari/537.36', 'https://vacation-sign-portal.ru/', '2026-06-04 08:58:00'),
-- Захаров: OPENED
('e5270921-2a1f-4bb0-8a19-482470eb0157','e5270921-2a1f-4bb0-8a19-482470eb0062','e5270921-2a1f-4bb0-8a19-482470eb0050','EMAIL_OPENED','10.10.4.55',  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0 Safari/537.36', NULL,                               '2026-06-05 11:40:00')
ON CONFLICT (id) DO NOTHING;
