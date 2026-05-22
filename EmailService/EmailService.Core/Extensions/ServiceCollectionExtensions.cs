using Microsoft.Extensions.DependencyInjection;

namespace EmailService.Core.Extensions;

/// <summary>
/// Provides extension methods for registering core email services.
/// </summary>
public static class ServiceCollectionExtensions
{
    /// <summary>
    /// Registers core application services.
    /// </summary>
    /// <param name="services">The service collection.</param>
    /// <returns>The updated service collection.</returns>
    public static IServiceCollection AddEmailCore(this IServiceCollection services)
    {
        return services;
    }
}