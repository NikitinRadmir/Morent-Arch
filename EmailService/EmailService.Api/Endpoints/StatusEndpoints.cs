using EmailService.Core.Contracts;
using EmailService.Core.Models;
using Microsoft.AspNetCore.Mvc;

namespace EmailService.Api.Endpoints;

public static class StatusEndpoints
{
    public static void MapStatusEndpoints(this IEndpointRouteBuilder app)
    {
        app.MapGet("/api/v1/emails/{correlationId}", async (
            string correlationId,
            IEmailStatusStore statusStore,
            CancellationToken ct) =>
        {
            var status = await statusStore.GetStatusAsync(correlationId, ct);

            return status.HasValue
                ? Results.Ok(new { correlationId, status = status.Value.ToString() })
                : Results.NotFound(new { error = "Request not found", correlationId });
        })
        .WithName("GetEmailStatus")
        .Produces(StatusCodes.Status200OK)
        .Produces(StatusCodes.Status404NotFound);
    }
}