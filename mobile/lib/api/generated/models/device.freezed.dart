// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'device.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Device {

 InstallationId get id; DevicePlatform get platform;@JsonKey(includeIfNull: true, name: 'app_version') String? get appVersion;@JsonKey(name: 'created_at') DateTime get createdAt;/// Last `registerDevice` of this installation; the 11th device evicts the oldest.
@JsonKey(name: 'last_seen_at') DateTime get lastSeenAt;
/// Create a copy of Device
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DeviceCopyWith<Device> get copyWith => _$DeviceCopyWithImpl<Device>(this as Device, _$identity);

  /// Serializes this Device to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Device;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Device&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.platform, _this.platform) || other.platform == _this.platform)&&(identical(other.appVersion, _this.appVersion) || other.appVersion == _this.appVersion)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.lastSeenAt, _this.lastSeenAt) || other.lastSeenAt == _this.lastSeenAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Device;
  return Object.hash(runtimeType,_this.id,_this.platform,_this.appVersion,_this.createdAt,_this.lastSeenAt);
}

@override
String toString() {
  final _this = this as Device;
  return 'Device(id: ${_this.id}, platform: ${_this.platform}, appVersion: ${_this.appVersion}, createdAt: ${_this.createdAt}, lastSeenAt: ${_this.lastSeenAt})';
}


}

/// @nodoc
abstract mixin class $DeviceCopyWith<$Res>  {
  factory $DeviceCopyWith(Device value, $Res Function(Device) _then) = _$DeviceCopyWithImpl;
@useResult
$Res call({
 InstallationId id, DevicePlatform platform,@JsonKey(includeIfNull: true, name: 'app_version') String? appVersion,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'last_seen_at') DateTime lastSeenAt
});




}
/// @nodoc
class _$DeviceCopyWithImpl<$Res>
    implements $DeviceCopyWith<$Res> {
  _$DeviceCopyWithImpl(this._self, this._then);

  final Device _self;
  final $Res Function(Device) _then;

/// Create a copy of Device
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? platform = null,Object? appVersion = freezed,Object? createdAt = null,Object? lastSeenAt = null,}) {
  return _then(Device(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as InstallationId,platform: null == platform ? _self.platform : platform // ignore: cast_nullable_to_non_nullable
as DevicePlatform,appVersion: freezed == appVersion ? _self.appVersion : appVersion // ignore: cast_nullable_to_non_nullable
as String?,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,lastSeenAt: null == lastSeenAt ? _self.lastSeenAt : lastSeenAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}

}


/// Adds pattern-matching-related methods to [Device].
extension DevicePatterns on Device {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Device value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Device() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Device value)  $default,){
final _that = this;
switch (_that) {
case _Device():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Device value)?  $default,){
final _that = this;
switch (_that) {
case _Device() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( InstallationId id,  DevicePlatform platform, @JsonKey(includeIfNull: true, name: 'app_version')  String? appVersion, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'last_seen_at')  DateTime lastSeenAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Device() when $default != null:
return $default(_that.id,_that.platform,_that.appVersion,_that.createdAt,_that.lastSeenAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( InstallationId id,  DevicePlatform platform, @JsonKey(includeIfNull: true, name: 'app_version')  String? appVersion, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'last_seen_at')  DateTime lastSeenAt)  $default,) {final _that = this;
switch (_that) {
case _Device():
return $default(_that.id,_that.platform,_that.appVersion,_that.createdAt,_that.lastSeenAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( InstallationId id,  DevicePlatform platform, @JsonKey(includeIfNull: true, name: 'app_version')  String? appVersion, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'last_seen_at')  DateTime lastSeenAt)?  $default,) {final _that = this;
switch (_that) {
case _Device() when $default != null:
return $default(_that.id,_that.platform,_that.appVersion,_that.createdAt,_that.lastSeenAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Device implements Device {
  const _Device({required this.id, required this.platform, @JsonKey(includeIfNull: true, name: 'app_version') required this.appVersion, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(name: 'last_seen_at') required this.lastSeenAt});
  factory _Device.fromJson(Map<String, dynamic> json) => _$DeviceFromJson(json);

@override final  InstallationId id;
@override final  DevicePlatform platform;
@override@JsonKey(includeIfNull: true, name: 'app_version') final  String? appVersion;
@override@JsonKey(name: 'created_at') final  DateTime createdAt;
/// Last `registerDevice` of this installation; the 11th device evicts the oldest.
@override@JsonKey(name: 'last_seen_at') final  DateTime lastSeenAt;

/// Create a copy of Device
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DeviceCopyWith<_Device> get copyWith => __$DeviceCopyWithImpl<_Device>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DeviceToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Device&&(identical(other.id, id) || other.id == id)&&(identical(other.platform, platform) || other.platform == platform)&&(identical(other.appVersion, appVersion) || other.appVersion == appVersion)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.lastSeenAt, lastSeenAt) || other.lastSeenAt == lastSeenAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,platform,appVersion,createdAt,lastSeenAt);
}

@override
String toString() {
    return 'Device(id: $id, platform: $platform, appVersion: $appVersion, createdAt: $createdAt, lastSeenAt: $lastSeenAt)';
}


}

/// @nodoc
abstract mixin class _$DeviceCopyWith<$Res> implements $DeviceCopyWith<$Res> {
  factory _$DeviceCopyWith(_Device value, $Res Function(_Device) _then) = __$DeviceCopyWithImpl;
@override @useResult
$Res call({
 InstallationId id, DevicePlatform platform,@JsonKey(includeIfNull: true, name: 'app_version') String? appVersion,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'last_seen_at') DateTime lastSeenAt
});




}
/// @nodoc
class __$DeviceCopyWithImpl<$Res>
    implements _$DeviceCopyWith<$Res> {
  __$DeviceCopyWithImpl(this._self, this._then);

  final _Device _self;
  final $Res Function(_Device) _then;

/// Create a copy of Device
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? platform = null,Object? appVersion = freezed,Object? createdAt = null,Object? lastSeenAt = null,}) {
  return _then(_Device(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as InstallationId,platform: null == platform ? _self.platform : platform // ignore: cast_nullable_to_non_nullable
as DevicePlatform,appVersion: freezed == appVersion ? _self.appVersion : appVersion // ignore: cast_nullable_to_non_nullable
as String?,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,lastSeenAt: null == lastSeenAt ? _self.lastSeenAt : lastSeenAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}


}

// dart format on
