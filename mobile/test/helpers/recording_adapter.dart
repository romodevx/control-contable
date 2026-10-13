import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:dio/dio.dart';

/// A canned reply for one request path.
typedef CannedReply = ({int status, Object? body});

/// A [HttpClientAdapter] that never touches the network: it records every
/// request and answers from its canned replies (by path), or `200 {}` by
/// default. A path mapped to an [Exception] makes the request fail with it.
class RecordingAdapter implements HttpClientAdapter {
  /// Creates the adapter with optional canned [replies] keyed by path.
  new({Map<String, Object> replies = const {}}) : _replies = Map.of(replies);

  final Map<String, Object> _replies;

  /// Every request that reached the adapter, in order.
  final List<RequestOptions> requests = [];

  /// Replaces the reply for [path] (a [CannedReply] or an [Exception]).
  void reply(String path, Object reply) => _replies[path] = reply;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    final reply =
        _replies[options.uri.path] ?? (status: 200, body: <String, Object?>{});
    if (reply is Exception) throw reply;
    final (:status, :body) = reply as CannedReply;
    return ResponseBody.fromString(
      jsonEncode(body),
      status,
      headers: {
        Headers.contentTypeHeader: [ContentType.json.mimeType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
