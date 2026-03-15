const std = @import("std");
const client_mod = @import("client.zig");
const json_build = @import("json_build.zig");
const json_parse = @import("json_parse.zig");
const base64 = @import("base64.zig");

pub const ImageConfig = struct {
    model: []const u8 = "gemini-2.0-flash-preview-image-generation",
    aspect_ratio: []const u8 = "16:9",
    image_size: []const u8 = "2K",
};

pub const GenerateResult = struct {
    image_data: []const u8, // decoded PNG bytes
    raw_json: []const u8, // full response body
    text: ?[]const u8, // optional text from response
};

const base_url = "https://generativelanguage.googleapis.com/v1beta/models/";

/// Build the full API URL for generateContent.
pub fn buildUrl(allocator: std.mem.Allocator, model: []const u8) ![]const u8 {
    return try std.fmt.allocPrint(allocator, "{s}{s}:generateContent", .{ base_url, model });
}

/// Generate an image via the Gemini API.
pub fn generate(
    allocator: std.mem.Allocator,
    http: *client_mod.HttpClient,
    api_key: []const u8,
    prompt: []const u8,
    cfg: ImageConfig,
) !GenerateResult {
    // Sanitize prompt
    const clean_prompt = try json_build.sanitizePrompt(allocator, prompt);

    // Build request
    const url = try buildUrl(allocator, cfg.model);
    const body = try json_build.buildRequest(allocator, clean_prompt, cfg.aspect_ratio, cfg.image_size);

    const headers = [_]client_mod.Header{
        .{ .name = "Content-Type", .value = "application/json" },
        .{ .name = "x-goog-api-key", .value = api_key },
    };

    // Send request
    const response = try http.post(url, &headers, body);

    if (response.status != .ok) {
        return error.ApiError;
    }

    // Parse response
    const parsed = try json_parse.extractImageData(allocator, response.body);

    // Decode base64 image
    const image_data = try base64.decode(allocator, parsed.image_base64);

    return GenerateResult{
        .image_data = image_data,
        .raw_json = response.body,
        .text = parsed.text,
    };
}

// Tests

test "buildUrl" {
    const allocator = std.testing.allocator;
    const url = try buildUrl(allocator, "gemini-3-pro-image-preview");
    defer allocator.free(url);
    try std.testing.expectEqualStrings(
        "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-pro-image-preview:generateContent",
        url,
    );
}

test "buildUrl default model" {
    const allocator = std.testing.allocator;
    const cfg = ImageConfig{};
    const url = try buildUrl(allocator, cfg.model);
    defer allocator.free(url);
    try std.testing.expect(std.mem.indexOf(u8, url, "gemini-2.0-flash-preview-image-generation") != null);
}

test "ImageConfig defaults" {
    const cfg = ImageConfig{};
    try std.testing.expectEqualStrings("gemini-2.0-flash-preview-image-generation", cfg.model);
    try std.testing.expectEqualStrings("16:9", cfg.aspect_ratio);
    try std.testing.expectEqualStrings("2K", cfg.image_size);
}
