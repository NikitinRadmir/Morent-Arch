using EmailService.Core.Models;
using FluentValidation;

namespace EmailService.Api.Validators;

public class EmailRequestValidator : AbstractValidator<EmailRequest>
{
    public EmailRequestValidator()
    {
        RuleFor(x => x.CorrelationId)
            .NotEmpty().WithMessage("CorrelationId is required")
            .Length(32, 36).WithMessage("CorrelationId must be a GUID");

        RuleFor(x => x.To)
            .NotEmpty().WithMessage("Email address is required")
            .EmailAddress().WithMessage("Invalid email format");

        RuleFor(x => x.TemplateKey)
            .NotEmpty().WithMessage("TemplateKey is required")
            .Matches(@"^[a-z0-9_\-]+$").WithMessage("TemplateKey must be lowercase alphanumeric with underscores/hyphens");

        RuleFor(x => x.Variables)
            .NotNull().WithMessage("Variables dictionary cannot be null");

        RuleFor(x => x.Priority)
            .IsInEnum();
    }
}