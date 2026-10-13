// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'event.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Event {

 String get id; String get type;@JsonKey(name: 'app_user_id') String get appUserId;@JsonKey(name: 'event_timestamp_ms') int get eventTimestampMs;@JsonKey(name: 'original_app_user_id') String? get originalAppUserId;@JsonKey(name: 'product_id') String? get productId;@JsonKey(name: 'entitlement_ids') List<String>? get entitlementIds; Environment? get environment; String? get store;@JsonKey(name: 'expiration_at_ms') int? get expirationAtMs;
/// Create a copy of Event
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$EventCopyWith<Event> get copyWith => _$EventCopyWithImpl<Event>(this as Event, _$identity);

  /// Serializes this Event to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Event;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Event&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.type, _this.type) || other.type == _this.type)&&(identical(other.appUserId, _this.appUserId) || other.appUserId == _this.appUserId)&&(identical(other.eventTimestampMs, _this.eventTimestampMs) || other.eventTimestampMs == _this.eventTimestampMs)&&(identical(other.originalAppUserId, _this.originalAppUserId) || other.originalAppUserId == _this.originalAppUserId)&&(identical(other.productId, _this.productId) || other.productId == _this.productId)&&const DeepCollectionEquality().equals(other.entitlementIds, _this.entitlementIds)&&(identical(other.environment, _this.environment) || other.environment == _this.environment)&&(identical(other.store, _this.store) || other.store == _this.store)&&(identical(other.expirationAtMs, _this.expirationAtMs) || other.expirationAtMs == _this.expirationAtMs));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Event;
  return Object.hash(runtimeType,_this.id,_this.type,_this.appUserId,_this.eventTimestampMs,_this.originalAppUserId,_this.productId,const DeepCollectionEquality().hash(_this.entitlementIds),_this.environment,_this.store,_this.expirationAtMs);
}

@override
String toString() {
  final _this = this as Event;
  return 'Event(id: ${_this.id}, type: ${_this.type}, appUserId: ${_this.appUserId}, eventTimestampMs: ${_this.eventTimestampMs}, originalAppUserId: ${_this.originalAppUserId}, productId: ${_this.productId}, entitlementIds: ${_this.entitlementIds}, environment: ${_this.environment}, store: ${_this.store}, expirationAtMs: ${_this.expirationAtMs})';
}


}

/// @nodoc
abstract mixin class $EventCopyWith<$Res>  {
  factory $EventCopyWith(Event value, $Res Function(Event) _then) = _$EventCopyWithImpl;
@useResult
$Res call({
 String id, String type,@JsonKey(name: 'app_user_id') String appUserId,@JsonKey(name: 'event_timestamp_ms') int eventTimestampMs,@JsonKey(name: 'original_app_user_id') String? originalAppUserId,@JsonKey(name: 'product_id') String? productId,@JsonKey(name: 'entitlement_ids') List<String>? entitlementIds, Environment? environment, String? store,@JsonKey(name: 'expiration_at_ms') int? expirationAtMs
});




}
/// @nodoc
class _$EventCopyWithImpl<$Res>
    implements $EventCopyWith<$Res> {
  _$EventCopyWithImpl(this._self, this._then);

  final Event _self;
  final $Res Function(Event) _then;

/// Create a copy of Event
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? type = null,Object? appUserId = null,Object? eventTimestampMs = null,Object? originalAppUserId = freezed,Object? productId = freezed,Object? entitlementIds = freezed,Object? environment = freezed,Object? store = freezed,Object? expirationAtMs = freezed,}) {
  return _then(Event(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,type: null == type ? _self.type : type // ignore: cast_nullable_to_non_nullable
as String,appUserId: null == appUserId ? _self.appUserId : appUserId // ignore: cast_nullable_to_non_nullable
as String,eventTimestampMs: null == eventTimestampMs ? _self.eventTimestampMs : eventTimestampMs // ignore: cast_nullable_to_non_nullable
as int,originalAppUserId: freezed == originalAppUserId ? _self.originalAppUserId : originalAppUserId // ignore: cast_nullable_to_non_nullable
as String?,productId: freezed == productId ? _self.productId : productId // ignore: cast_nullable_to_non_nullable
as String?,entitlementIds: freezed == entitlementIds ? _self.entitlementIds : entitlementIds // ignore: cast_nullable_to_non_nullable
as List<String>?,environment: freezed == environment ? _self.environment : environment // ignore: cast_nullable_to_non_nullable
as Environment?,store: freezed == store ? _self.store : store // ignore: cast_nullable_to_non_nullable
as String?,expirationAtMs: freezed == expirationAtMs ? _self.expirationAtMs : expirationAtMs // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}

}


/// Adds pattern-matching-related methods to [Event].
extension EventPatterns on Event {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Event value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Event() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Event value)  $default,){
final _that = this;
switch (_that) {
case _Event():
return $default(_that);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Event value)?  $default,){
final _that = this;
switch (_that) {
case _Event() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String id,  String type, @JsonKey(name: 'app_user_id')  String appUserId, @JsonKey(name: 'event_timestamp_ms')  int eventTimestampMs, @JsonKey(name: 'original_app_user_id')  String? originalAppUserId, @JsonKey(name: 'product_id')  String? productId, @JsonKey(name: 'entitlement_ids')  List<String>? entitlementIds,  Environment? environment,  String? store, @JsonKey(name: 'expiration_at_ms')  int? expirationAtMs)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Event() when $default != null:
return $default(_that.id,_that.type,_that.appUserId,_that.eventTimestampMs,_that.originalAppUserId,_that.productId,_that.entitlementIds,_that.environment,_that.store,_that.expirationAtMs);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String id,  String type, @JsonKey(name: 'app_user_id')  String appUserId, @JsonKey(name: 'event_timestamp_ms')  int eventTimestampMs, @JsonKey(name: 'original_app_user_id')  String? originalAppUserId, @JsonKey(name: 'product_id')  String? productId, @JsonKey(name: 'entitlement_ids')  List<String>? entitlementIds,  Environment? environment,  String? store, @JsonKey(name: 'expiration_at_ms')  int? expirationAtMs)  $default,) {final _that = this;
switch (_that) {
case _Event():
return $default(_that.id,_that.type,_that.appUserId,_that.eventTimestampMs,_that.originalAppUserId,_that.productId,_that.entitlementIds,_that.environment,_that.store,_that.expirationAtMs);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String id,  String type, @JsonKey(name: 'app_user_id')  String appUserId, @JsonKey(name: 'event_timestamp_ms')  int eventTimestampMs, @JsonKey(name: 'original_app_user_id')  String? originalAppUserId, @JsonKey(name: 'product_id')  String? productId, @JsonKey(name: 'entitlement_ids')  List<String>? entitlementIds,  Environment? environment,  String? store, @JsonKey(name: 'expiration_at_ms')  int? expirationAtMs)?  $default,) {final _that = this;
switch (_that) {
case _Event() when $default != null:
return $default(_that.id,_that.type,_that.appUserId,_that.eventTimestampMs,_that.originalAppUserId,_that.productId,_that.entitlementIds,_that.environment,_that.store,_that.expirationAtMs);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Event implements Event {
  const _Event({required this.id, required this.type, @JsonKey(name: 'app_user_id') required this.appUserId, @JsonKey(name: 'event_timestamp_ms') required this.eventTimestampMs, @JsonKey(name: 'original_app_user_id') this.originalAppUserId, @JsonKey(name: 'product_id') this.productId, @JsonKey(name: 'entitlement_ids')  List<String>? entitlementIds, this.environment, this.store, @JsonKey(name: 'expiration_at_ms') this.expirationAtMs}): _entitlementIds = entitlementIds;
  factory _Event.fromJson(Map<String, dynamic> json) => _$EventFromJson(json);

@override final  String id;
@override final  String type;
@override@JsonKey(name: 'app_user_id') final  String appUserId;
@override@JsonKey(name: 'event_timestamp_ms') final  int eventTimestampMs;
@override@JsonKey(name: 'original_app_user_id') final  String? originalAppUserId;
@override@JsonKey(name: 'product_id') final  String? productId;
 final  List<String>? _entitlementIds;
@override@JsonKey(name: 'entitlement_ids') List<String>? get entitlementIds {
  final value = _entitlementIds;
  if (value == null) return null;
  if (_entitlementIds is EqualUnmodifiableListView) return _entitlementIds;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(value);
}

@override final  Environment? environment;
@override final  String? store;
@override@JsonKey(name: 'expiration_at_ms') final  int? expirationAtMs;

/// Create a copy of Event
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$EventCopyWith<_Event> get copyWith => __$EventCopyWithImpl<_Event>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$EventToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Event&&(identical(other.id, id) || other.id == id)&&(identical(other.type, type) || other.type == type)&&(identical(other.appUserId, appUserId) || other.appUserId == appUserId)&&(identical(other.eventTimestampMs, eventTimestampMs) || other.eventTimestampMs == eventTimestampMs)&&(identical(other.originalAppUserId, originalAppUserId) || other.originalAppUserId == originalAppUserId)&&(identical(other.productId, productId) || other.productId == productId)&&const DeepCollectionEquality().equals(other.entitlementIds, _entitlementIds)&&(identical(other.environment, environment) || other.environment == environment)&&(identical(other.store, store) || other.store == store)&&(identical(other.expirationAtMs, expirationAtMs) || other.expirationAtMs == expirationAtMs));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,type,appUserId,eventTimestampMs,originalAppUserId,productId,const DeepCollectionEquality().hash(_entitlementIds),environment,store,expirationAtMs);
}

@override
String toString() {
    return 'Event(id: $id, type: $type, appUserId: $appUserId, eventTimestampMs: $eventTimestampMs, originalAppUserId: $originalAppUserId, productId: $productId, entitlementIds: $entitlementIds, environment: $environment, store: $store, expirationAtMs: $expirationAtMs)';
}


}

/// @nodoc
abstract mixin class _$EventCopyWith<$Res> implements $EventCopyWith<$Res> {
  factory _$EventCopyWith(_Event value, $Res Function(_Event) _then) = __$EventCopyWithImpl;
@override @useResult
$Res call({
 String id, String type,@JsonKey(name: 'app_user_id') String appUserId,@JsonKey(name: 'event_timestamp_ms') int eventTimestampMs,@JsonKey(name: 'original_app_user_id') String? originalAppUserId,@JsonKey(name: 'product_id') String? productId,@JsonKey(name: 'entitlement_ids') List<String>? entitlementIds, Environment? environment, String? store,@JsonKey(name: 'expiration_at_ms') int? expirationAtMs
});




}
/// @nodoc
class __$EventCopyWithImpl<$Res>
    implements _$EventCopyWith<$Res> {
  __$EventCopyWithImpl(this._self, this._then);

  final _Event _self;
  final $Res Function(_Event) _then;

/// Create a copy of Event
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? type = null,Object? appUserId = null,Object? eventTimestampMs = null,Object? originalAppUserId = freezed,Object? productId = freezed,Object? entitlementIds = freezed,Object? environment = freezed,Object? store = freezed,Object? expirationAtMs = freezed,}) {
  return _then(_Event(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as String,type: null == type ? _self.type : type // ignore: cast_nullable_to_non_nullable
as String,appUserId: null == appUserId ? _self.appUserId : appUserId // ignore: cast_nullable_to_non_nullable
as String,eventTimestampMs: null == eventTimestampMs ? _self.eventTimestampMs : eventTimestampMs // ignore: cast_nullable_to_non_nullable
as int,originalAppUserId: freezed == originalAppUserId ? _self.originalAppUserId : originalAppUserId // ignore: cast_nullable_to_non_nullable
as String?,productId: freezed == productId ? _self.productId : productId // ignore: cast_nullable_to_non_nullable
as String?,entitlementIds: freezed == entitlementIds ? _self._entitlementIds : entitlementIds // ignore: cast_nullable_to_non_nullable
as List<String>?,environment: freezed == environment ? _self.environment : environment // ignore: cast_nullable_to_non_nullable
as Environment?,store: freezed == store ? _self.store : store // ignore: cast_nullable_to_non_nullable
as String?,expirationAtMs: freezed == expirationAtMs ? _self.expirationAtMs : expirationAtMs // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}


}

// dart format on
