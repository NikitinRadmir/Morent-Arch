using EmailService.Core.Contracts;
using EmailService.Core.Models;
using MassTransit;
using Microsoft.AspNetCore.Mvc;
using Serilog;

namespace EmailService.Api.Endpoints;

/// <summary>
/// Contains minimal API endpoints for email ingestion.
/// </summary>
public static class IngestionEndpoints
{
    /// <summary>
    /// Maps email ingestion endpoints to the application pipeline.
    /// </summary>
    /// <param name="app">The endpoint route builder.</param>
    public static void MapIngestionEndpoints(this IEndpointRouteBuilder app)
    {
        app.MapPost("/api/v1/emails", async (
            EmailRequest request,
            IPublishEndpoint bus,
            IEmailStatusStore statusStore,
            CancellationToken ct) =>
        {
            using (Serilog.Context.LogContext.PushProperty("CorrelationId", request.CorrelationId))
            {
                Log.Information(
                    "Received email request: Template={TemplateKey}, To={To}, Priority={Priority}",
                    request.TemplateKey, request.To, request.Priority);

                try
                {
                    await statusStore.MarkStatusAsync(request.CorrelationId, EmailStatus.Queued, ct: ct);
                    Log.Debug("Status marked as Queued in database");

                    await bus.Publish(request, ct);
                    Log.Information("Published to RabbitMQ successfully");

                    Log.Information("Request processed successfully");
                    return Results.Accepted(
                        $"/api/v1/emails/{request.CorrelationId}",
                        new { correlationId = request.CorrelationId, status = "queued" });
                }
                catch (Exception ex)
                {
                    Log.Error(ex, "Failed to process email request");
                    return Results.Problem(
                        detail: "Failed to queue email",
                        statusCode: StatusCodes.Status500InternalServerError);
                }
            }
        })
        .WithName("SendEmail")
        .Produces(StatusCodes.Status202Accepted)
        .Produces(StatusCodes.Status400BadRequest)
        .Produces(StatusCodes.Status500InternalServerError);
    }
}