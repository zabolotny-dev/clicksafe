export const defaultMessageTemplate = `<main style="font-family: Arial, sans-serif; max-width: 680px; margin: 0 auto; color: #18202b;">
  <h1 style="color: #157f72;">Привет, {{ .Employee.FirstName }}!</h1>
  <p>Это демо-письмо для сотрудника {{ .Employee.FirstName }} {{ .Employee.LastName }}.</p>
  <p><b>Отдел:</b> {{ .Department.Label }}</p>
  <p><b>Компания:</b> {{ .Organization.Label }}</p>
  <p><b>Email:</b> {{ .Employee.Email.Address }}</p>
  <p><b>Роль:</b> {{ .Employee.Attributes.Role }}</p>
  <p><a href="{{ .Target.Link }}" style="display: inline-block; padding: 12px 16px; background: #157f72; color: white; text-decoration: none; border-radius: 6px;">Открыть учебный портал</a></p>
</main>`

export const defaultLandingTemplate = `<main style="font-family: Arial, sans-serif; max-width: 720px; margin: 0 auto; padding: 40px 24px; color: #18202b;">
  <h1 style="color: #157f72;">Учебный портал {{ .Organization.Label }}</h1>
  <p>Здравствуйте, {{ .Employee.FirstName }} {{ .Employee.LastName }}.</p>
  <p>Для отдела {{ .Department.Label }} доступна проверочная страница обучения.</p>
  <form style="display: grid; gap: 12px; max-width: 360px; margin-top: 24px;">
    <input placeholder="Email" value="{{ .Employee.Email.Address }}" style="padding: 12px; border: 1px solid #d9e0e7; border-radius: 6px;" />
    <input placeholder="Password" type="password" style="padding: 12px; border: 1px solid #d9e0e7; border-radius: 6px;" />
    <button type="button" style="padding: 12px; border: 0; border-radius: 6px; color: white; background: #157f72;">Продолжить</button>
  </form>
</main>`

export const defaultTemplate = defaultMessageTemplate
