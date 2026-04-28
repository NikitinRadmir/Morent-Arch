using Microsoft.Extensions.DependencyInjection;

namespace EmailService.Core.Extensions;

public static class ServiceCollectionExtensions
{
    public static IServiceCollection AddEmailCore(this IServiceCollection services)
    {

        return services;
    }
}