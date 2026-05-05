namespace EmailService.Core.Models;

public enum EmailStatus
{
    Queued,      // Принято в систему
    Rendering,   // Шаблон обрабатывается
    Sent,        // Передано провайдеру
    Delivered,   // Доставлено в почтовый ящик
    Bounced,     // Возвращено (невалидный адрес/ящик полон)
    Failed,      // Ошибка отправки после ретраев
    Spam,        // Пользователь пожаловался
    Unsubscribed // Отписка
}