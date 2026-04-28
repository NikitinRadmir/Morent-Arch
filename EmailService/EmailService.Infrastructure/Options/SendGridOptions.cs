namespace EmailService.Infrastructure.Options;

public class SendGridOptions
{
    public string ApiKey { get; set; } = null!;
    public string FromEmail { get; set; } = null!;
    public string FromName { get; set; } = "AutoRental";
}