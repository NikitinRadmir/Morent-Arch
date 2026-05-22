using EmailService.Core.Contracts;
using EmailService.Core.Models;
using Microsoft.AspNetCore.Mvc;
using Serilog;
using Serilog.Context;

namespace EmailService.Api.Endpoints;

/// <summary>
/// Endpoints for retrieving email processing status.
/// </summary>
public static class StatusEndpoints
{
    /// <summary>
    /// Maps status-related endpoints to the application pipeline.
    /// </summary>
    /// <param name="app">Endpoint route builder.</param>
    public static void MapStatusEndpoints(this IEndpointRouteBuilder app)
    {
        app.MapGet("/api/v1/emails/{correlationId}", async (
            string correlationId,
            IEmailStatusStore statusStore,
            CancellationToken ct) =>
        {
            using (LogContext.PushProperty("CorrelationId", correlationId))
            {
                Log.Debug("Checking status for correlationId={CorrelationId}", correlationId);

                try
                {
                    var status = await statusStore.GetStatusAsync(correlationId, ct);

                    if (status.HasValue)
                    {
                        Log.Information("Status found: {Status}", status.Value);
                        return Results.Ok(new { correlationId, status = status.Value.ToString() });
                    }

                    Log.Warning("Status not found for correlationId={CorrelationId}", correlationId);
                    return Results.NotFound(new { error = "Request not found", correlationId });
                }
                catch (Exception ex)
                {
                    Log.Error(ex, "Failed to retrieve status for {CorrelationId}", correlationId);
                    return Results.Problem(
                        detail: "Failed to retrieve status",
                        statusCode: StatusCodes.Status500InternalServerError);
                }
            }
        })
        .WithName("GetEmailStatus")
        .Produces(StatusCodes.Status200OK)
        .Produces(StatusCodes.Status404NotFound)
        .Produces(StatusCodes.Status500InternalServerError);
    }
}