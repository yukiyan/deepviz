const std = @import("std");
const cli = @import("cli.zig");
const config_mod = @import("config.zig");
const fs = @import("fs.zig");
const api = @import("api.zig");
const client_mod = @import("client.zig");
const opener = @import("opener.zig");
const log_mod = @import("log.zig");
const timestamp_mod = @import("timestamp.zig");

const version = "0.1.0";

pub fn main() !void {
    run() catch |e| {
        const stderr = std.io.getStdErr().writer();
        stderr.print("error: {s}\n", .{@errorName(e)}) catch {};
        std.process.exit(1);
    };
}

fn run() !void {
    // Arena allocator for entire program lifetime
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    const allocator = arena.allocator();

    // Parse CLI args (skip argv[0] = program name)
    const argv = try std.process.argsAlloc(allocator);
    const args = cli.parse(argv[1..]) catch {
        try cli.printUsage(std.io.getStdErr().writer());
        std.process.exit(1);
        unreachable;
    };

    // Handle --help
    if (args.help) {
        try cli.printUsage(std.io.getStdOut().writer());
        return;
    }

    // Handle --version
    if (args.version) {
        const stdout = std.io.getStdOut().writer();
        try stdout.print("nanogen {s}\n", .{version});
        return;
    }

    // Load config
    const cfg = try config_mod.load(allocator, args);

    // Generate timestamp
    const ts = timestamp_mod.generate();
    const ts_str: []const u8 = &ts;

    // Setup logger
    const log_dir = try config_mod.logsDir(allocator, cfg.output_dir);
    const log_path = try std.fmt.allocPrint(allocator, "{s}/{s}.log", .{ log_dir, ts_str });
    var logger = log_mod.Logger.init(cfg.verbose, log_path);
    defer logger.deinit();

    // Validate: need either --prompt or --file
    if (args.prompt == null and args.file == null) {
        const stderr = std.io.getStdErr().writer();
        try stderr.writeAll("error: either --prompt or --file must be specified\n\n");
        try cli.printUsage(stderr);
        std.process.exit(1);
    }

    // Validate API key
    if (cfg.api_key.len == 0) {
        const stderr = std.io.getStdErr().writer();
        try stderr.writeAll("error: API key not set. Set GEMINI_API_KEY or NANOGEN_API_KEY environment variable\n");
        std.process.exit(1);
    }

    // Get prompt text
    var prompt: []const u8 = undefined;
    if (args.file) |file_path| {
        prompt = try fs.readFileAlloc(allocator, file_path);
        if (prompt.len == 0) {
            const stderr = std.io.getStdErr().writer();
            try stderr.print("error: prompt file is empty: {s}\n", .{file_path});
            std.process.exit(1);
        }
        logger.info("Loaded prompt from file: {s}", .{file_path});
    } else {
        prompt = args.prompt.?;
    }

    logger.info("Starting image generation", .{});
    logger.info("Model: {s} | Aspect: {s} | Size: {s}", .{ cfg.model, cfg.aspect_ratio, cfg.image_size });

    // Create HTTP client and generate
    var http = client_mod.HttpClient.init(allocator);
    defer http.deinit();

    const img_config = api.ImageConfig{
        .model = cfg.model,
        .aspect_ratio = cfg.aspect_ratio,
        .image_size = cfg.image_size,
    };

    const result = try api.generate(allocator, &http, cfg.api_key, prompt, img_config);

    // Ensure output directories
    const images_dir = try config_mod.imagesDir(allocator, cfg.output_dir);
    const responses_dir = try config_mod.responsesDir(allocator, cfg.output_dir);
    try fs.ensureDir(images_dir);
    try fs.ensureDir(responses_dir);

    // Save files
    const image_path = try std.fmt.allocPrint(allocator, "{s}/{s}.png", .{ images_dir, ts_str });
    const response_path = try std.fmt.allocPrint(allocator, "{s}/{s}_image.json", .{ responses_dir, ts_str });

    try fs.writeFile(image_path, result.image_data);
    logger.info("Image saved: {s}", .{image_path});

    try fs.writeFile(response_path, result.raw_json);
    logger.info("Response saved: {s}", .{response_path});

    // Auto-open
    if (cfg.auto_open) {
        opener.openFile(allocator, image_path) catch |e| {
            logger.info("Failed to open image: {s}", .{@errorName(e)});
        };
    }

    // Print summary
    const stdout = std.io.getStdOut().writer();
    try stdout.print("\n=== Generation Complete ===\n", .{});
    try stdout.print("Image: {s}\n", .{image_path});
    try stdout.print("Output: {s}\n", .{cfg.output_dir});
}

// Tests for main module

test "version string" {
    try std.testing.expectEqualStrings("0.1.0", version);
}

test "imports compile" {
    // Verify all imports resolve
    _ = cli.Args;
    _ = config_mod.Config;
    _ = api.ImageConfig;
    _ = client_mod.HttpClient;
    _ = log_mod.Logger;
    _ = timestamp_mod.generate;
}
