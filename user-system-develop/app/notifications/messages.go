package notifications

import (
	"fmt"
	"time"
)

type MessageBuilder struct {
	baseLink string
}

func NewMessageBuilder(baseLink string) *MessageBuilder {
	return &MessageBuilder{
		baseLink: baseLink,
	}
}

func (b *MessageBuilder) BuildWelcomeMessage(firstName, companyName string) (subject, message string) {
	subject = fmt.Sprintf("Добро пожаловать в %s!", companyName)

	message = fmt.Sprintf(`Здравствуйте, %s!

	Вы успешно зарегистрировались в системе %s.
	
	С уважением,
	Команда %s`, firstName, companyName, companyName)

	return subject, message
}

func (b *MessageBuilder) BuildLoginMessage(firstName, companyName, ipAddress string) (subject, message string) {
	subject = "Новый вход в аккаунт"

	currentTime := time.Now().Format("02.01.2006 15:04:05")

	message = fmt.Sprintf(`Здравствуйте, %s!

	Зафиксирован вход в ваш аккаунт в системе %s.
	
	Время: %s
	IP адрес: %s
	
	Если это были не вы, обратитесь в поддержку.
	
	С уважением,
	Команда %s`,
		firstName, companyName, currentTime, ipAddress, companyName)

	return subject, message
}
