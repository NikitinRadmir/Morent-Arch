using EmailService.Core.Contracts;
using EmailService.Core.Models;
using EmailService.Infrastructure.Security;
using Microsoft.AspNetCore.Mvc;

namespace EmailService.Api.Endpoints;

public static class WebhookEndpoints
{
    public static void MapWebhookEndpoints(this IEndpointRouteBuilder app)
    {
        app.MapPost("/webhooks/sendgrid", async (
            HttpContext ctx,
            IEmailStatusStore statusStore,
            WebhookHmacValidator validator,
            CancellationToken ct) =>
        {
            var events = await ctx.Request.ReadFromJsonAsync<SendGridWebhookEvent[]>(ct);

            if (!ctx.Request.Headers.TryGetValue("X-Sendgrid-Signature", out var signature) ||
                !validator.Verify(signature, events))
            {
                return Results.Unauthorized();
            }

            _ = Task.Run(async () =>
            {
                if (events == null) return;

                foreach (var evt in events)
                {
                    EmailStatus? status = evt.EventType switch
                    {
                        "delivered" => EmailStatus.Delivered,
                        "bounce" => EmailStatus.Bounced,
                        "spamreport" => EmailStatus.Spam,
                        "unsubscribe" => EmailStatus.Unsubscribed,
                        "dropped" or "deferred" or "processed" => null,
                        _ => null
                    };

                    if (status.HasValue && !string.IsNullOrEmpty(evt.CorrelationId))
                    {
                        await statusStore.MarkStatusAsync(evt.CorrelationId, status.Value,
                            evt.Reason, ct);
                    }
                }
            }, ct);

            return Results.Ok();
        })
        .AllowAnonymous()
        .DisableAntiforgery();
    }
}

public record SendGridWebhookEvent
{
    public string Email { get; init; } = null!;
    public string EventType { get; init; } = null!;
    public string? Reason { get; init; }
    public string? CorrelationId { get; init; }
    public string? ProviderMessageId { get; init; }
}