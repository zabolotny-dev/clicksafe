export const defaultTemplate = `<main style="font-family: Arial, sans-serif; max-width: 680px; margin: 0 auto; color: #18202b;">
  <h1 style="color: #157f72;">Привет, {{ .Employee.FirstName }}!</h1>
  <p>Это демо-письмо для сотрудника {{ .Employee.FirstName }} {{ .Employee.LastName }}.</p>
  <p><b>Отдел:</b> {{ .Department.Label }}</p>
  <p><b>Компания:</b> {{ .Organization.Label }}</p>
  <p><b>Email:</b> {{ .Employee.Email.Address }}</p>
  <p><b>Роль:</b> {{ .Employee.Attributes.Role }}</p>
</main>`
