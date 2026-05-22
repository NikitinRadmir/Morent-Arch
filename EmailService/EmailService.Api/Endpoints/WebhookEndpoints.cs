using EmailService.Core.Contracts;
using EmailService.Core.Models;
using Microsoft.AspNetCore.Mvc;
using Serilog;

namespace EmailService.Api.Endpoints;

/// <summary>
/// Endpoints for processing webhook events from email providers.
/// </summary>
public static class WebhookEndpoints
{
    /// <summary>
    /// Maps webhook-related endpoints to the application pipeline.
    /// </summary>
    /// <param name="app">Endpoint route builder.</param>
    public static void MapWebhookEndpoints(this IEndpointRouteBuilder app)
    {
        app.MapPost("/webhooks/resend", async (
            [FromBody] ResendWebhookEvent webhookEvent,
            IEmailStatusStore statusStore,
            CancellationToken ct) =>
        {
            Log.Debug("Received webhook from Resend");

            try
            {
                if (webhookEvent == null)
                {
                    Log.Warning("Webhook payload is null or invalid");
                    return Results.BadRequest();
                }

                var correlationId = webhookEvent.CorrelationId;

                if (string.IsNullOrEmpty(correlationId))
                {
                    Log.Warning("Webhook received without CorrelationId");
                    return Results.Ok();
                }

                using (Serilog.Context.LogContext.PushProperty("CorrelationId", correlationId))
                {
                    Log.Information(
                        "Processing webhook: Event={Event}, Email={Email}",
                        webhookEvent.Event, webhookEvent.Email);

                    EmailStatus? status = webhookEvent.Event switch
                    {
                        "sent" => EmailStatus.Sent,
                        "delivered" => EmailStatus.Delivered,
                        "complained" => EmailStatus.Spam,
                        "bounced" => EmailStatus.Bounced,
                        "failed" or "dropped" => EmailStatus.Failed,
                        _ => null
                    };

                    if (status.HasValue)
                    {
                        await statusStore.MarkStatusAsync(
                            correlationId,
                            status.Value,
                            webhookEvent.Description,
                            ct);

                        Log.Information("Status updated to {Status} for {CorrelationId}", status.Value, correlationId);
                    }
                    else
                    {
                        Log.Debug("Event {Event} has no status mapping, skipping", webhookEvent.Event);
                    }
                }

                return Results.Ok();
            }
            catch (Exception ex)
            {
                Log.Error(ex, "Failed to process webhook");
                return Results.Ok();
            }
        })
        .WithName("ResendWebhook")
        .Produces(StatusCodes.Status200OK)
        .Produces(StatusCodes.Status400BadRequest)
        .AllowAnonymous()
        .DisableAntiforgery();
    }
}

/// <summary>
/// Represents a webhook event received from Resend email service.
/// </summary>
public record ResendWebhookEvent
{
    /// <summary>
    /// Type of the event (e.g. email.sent, email.delivered).
    /// </summary>
    public string Type { get; init; } = null!;

    /// <summary>
    /// Event status (sent, delivered, bounced, complained, failed, dropped).
    /// </summary>
    public string Event { get; init; } = null!;

    /// <summary>
    /// Correlation identifier used to track the email request.
    /// </summary>
    public string? CorrelationId { get; init; }

    /// <summary>
    /// Optional description or error message related to the event.
    /// </summary>
    public string? Description { get; init; }

    /// <summary>
    /// Recipient email address.
    /// </summary>
    public string? Email { get; init; }

    /// <summary>
    /// Timestamp when the event occurred.
    /// </summary>
    public string? CreatedAt { get; init; }

    /// <summary>
    /// Provider-specific event identifier.
    /// </summary>
    public string? Id { get; init; }
}