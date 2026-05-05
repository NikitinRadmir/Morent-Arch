using System.Threading.Channels;
using EmailService.Core.Contracts;
using EmailService.Core.Models;

namespace EmailService.Infrastructure.Queue;

public class ChannelEmailQueue : IEmailQueue
{
    private readonly Channel<EmailRequest> _channel;

    public ChannelEmailQueue(int capacity = 1000)
    {
        _channel = Channel.CreateBounded<EmailRequest>(new BoundedChannelOptions(capacity)
        {
            FullMode = BoundedChannelFullMode.Wait,
            SingleWriter = false,
            SingleReader = false
        });
    }

    public ValueTask EnqueueAsync(EmailRequest request, CancellationToken ct = default) =>
        _channel.Writer.WriteAsync(request, ct);

    public IAsyncEnumerable<EmailRequest> ReadAllAsync(CancellationToken ct = default) =>
        _channel.Reader.ReadAllAsync(ct);
}