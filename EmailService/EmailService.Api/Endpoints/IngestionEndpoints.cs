using EmailService.Core.Contracts;
using EmailService.Core.Models;
using Microsoft.AspNetCore.Mvc;

namespace EmailService.Api.Endpoints;

public static class IngestionEndpoints
{
    public static void MapIngestionEndpoints(this IEndpointRouteBuilder app)
    {
        app.MapPost("/api/v1/emails", async (
            EmailRequest request,
            IEmailQueue queue,
            IEmailStatusStore statusStore,
            CancellationToken ct) =>
        {
            if (await statusStore.IsProcessedAsync(request.CorrelationId, ct))
            {
                return Results.Ok(new { status = "already_processed", correlationId = request.CorrelationId });
            }

            await statusStore.MarkStatusAsync(request.CorrelationId, EmailStatus.Queued, ct: ct);

            await queue.EnqueueAsync(request, ct);

            return Results.Accepted($"/api/v1/emails/{request.CorrelationId}",
                new { correlationId = request.CorrelationId, status = "queued" });
        })
        .WithName("SendEmail")
        .Produces(StatusCodes.Status202Accepted)
        .Produces(StatusCodes.Status200OK)
        .Produces(StatusCodes.Status400BadRequest)
        .Produces(StatusCodes.Status500InternalServerError);
    }
}