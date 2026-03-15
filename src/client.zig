const std = @import("std");

pub const Header = struct {
    name: []const u8,
    value: []const u8,
};

pub const Response = struct {
    status: std.http.Status,
    body: []const u8,
};

pub const HttpClient = struct {
    inner: std.http.Client,
    allocator: std.mem.Allocator,

    pub fn init(allocator: std.mem.Allocator) HttpClient {
        return .{
            .inner = std.http.Client{ .allocator = allocator },
            .allocator = allocator,
        };
    }

    pub fn deinit(self: *HttpClient) void {
        self.inner.deinit();
    }

    /// POST request. Caller owns returned body via allocator.
    pub fn post(self: *HttpClient, url_str: []const u8, headers: []const Header, body: []const u8) !Response {
        const uri = try std.Uri.parse(url_str);

        var extra_headers = std.http.Client.Request.Headers{};
        for (headers) |h| {
            if (std.mem.eql(u8, h.name, "Content-Type")) {
                extra_headers.content_type = .{ .override = h.value };
            }
        }

        var req = try self.inner.open(.POST, uri, .{
            .server_header_buffer = try self.allocator.alloc(u8, 16 * 1024),
            .extra_headers = blk: {
                var list = std.ArrayList(std.http.Header).init(self.allocator);
                for (headers) |h| {
                    if (!std.mem.eql(u8, h.name, "Content-Type")) {
                        try list.append(.{ .name = h.name, .value = h.value });
                    }
                }
                break :blk try list.toOwnedSlice();
            },
        });
        defer req.deinit();

        req.transfer_encoding = .{ .content_length = body.len };
        try req.send();
        try req.writer().writeAll(body);
        try req.finish();
        try req.wait();

        // Read response body
        const max_size = 32 * 1024 * 1024; // 32MB max for image responses
        const resp_body = try req.reader().readAllAlloc(self.allocator, max_size);

        return Response{
            .status = req.response.status,
            .body = resp_body,
        };
    }
};

test "HttpClient init/deinit" {
    var client = HttpClient.init(std.testing.allocator);
    client.deinit();
}

test "Response struct" {
    const resp = Response{
        .status = .ok,
        .body = "test",
    };
    try std.testing.expectEqual(std.http.Status.ok, resp.status);
    try std.testing.expectEqualStrings("test", resp.body);
}
