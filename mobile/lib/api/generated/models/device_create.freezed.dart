// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'device_create.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$DeviceCreate {

 Uuid get id; DeviceCreatePlatform get platform;@JsonKey(name: 'push_token') String get pushToken;@JsonKey(name: 'app_version') String? get appVersion;
/// Create a copy of DeviceCreate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DeviceCreateCopyWith<DeviceCreate> get copyWith => _$DeviceCreateCopyWithImpl<DeviceCreate>(this as DeviceCreate, _$identity);

  /// Serializes this DeviceCreate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as DeviceCreate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is DeviceCreate&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.platform, _this.platform) || other.platform == _this.platform)&&(identical(other.pushToken, _this.pushToken) || other.pushToken == _this.pushToken)&&(identical(other.appVersion, _this.appVersion) || other.appVersion == _this.appVersion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as DeviceCreate;
  return Object.hash(runtimeType,_this.id,_this.platform,_this.pushToken,_this.appVersion);
}

@override
String toString() {
  final _this = this as DeviceCreate;
  return 'DeviceCreate(id: ${_this.id}, platform: ${_this.platform}, pushToken: ${_this.pushToken}, appVersion: ${_this.appVersion})';
}


}

/// @nodoc
abstract mixin class $DeviceCreateCopyWith<$Res>  {
  factory $DeviceCreateCopyWith(DeviceCreate value, $Res Function(DeviceCreate) _then) = _$DeviceCreateCopyWithImpl;
@useResult
$Res call({
 Uuid id, DeviceCreatePlatform platform,@JsonKey(name: 'push_token') String pushToken,@JsonKey(name: 'app_version') String? appVersion
});




}
/// @nodoc
class _$DeviceCreateCopyWithImpl<$Res>
    implements $DeviceCreateCopyWith<$Res> {
  _$DeviceCreateCopyWithImpl(this._self, this._then);

  final DeviceCreate _self;
  final $Res Function(DeviceCreate) _then;

/// Create a copy of DeviceCreate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? platform = null,Object? pushToken = null,Object? appVersion = freezed,}) {
  return _then(DeviceCreate(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,platform: null == platform ? _self.platform : platform // ignore: cast_nullable_to_non_nullable
as DeviceCreatePlatform,pushToken: null == pushToken ? _self.pushToken : pushToken // ignore: cast_nullable_to_non_nullable
as String,appVersion: freezed == appVersion ? _self.appVersion : appVersion // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}

}


/// Adds pattern-matching-related methods to [DeviceCreate].
extension DeviceCreatePatterns on DeviceCreate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _DeviceCreate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _DeviceCreate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _DeviceCreate value)  $default,){
final _that = this;
switch (_that) {
case _DeviceCreate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _DeviceCreate value)?  $default,){
final _that = this;
switch (_that) {
case _DeviceCreate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Uuid id,  DeviceCreatePlatform platform, @JsonKey(name: 'push_token')  String pushToken, @JsonKey(name: 'app_version')  String? appVersion)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _DeviceCreate() when $default != null:
return $default(_that.id,_that.platform,_that.pushToken,_that.appVersion);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Uuid id,  DeviceCreatePlatform platform, @JsonKey(name: 'push_token')  String pushToken, @JsonKey(name: 'app_version')  String? appVersion)  $default,) {final _that = this;
switch (_that) {
case _DeviceCreate():
return $default(_that.id,_that.platform,_that.pushToken,_that.appVersion);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Uuid id,  DeviceCreatePlatform platform, @JsonKey(name: 'push_token')  String pushToken, @JsonKey(name: 'app_version')  String? appVersion)?  $default,) {final _that = this;
switch (_that) {
case _DeviceCreate() when $default != null:
return $default(_that.id,_that.platform,_that.pushToken,_that.appVersion);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _DeviceCreate implements DeviceCreate {
  const _DeviceCreate({required this.id, required this.platform, @JsonKey(name: 'push_token') required this.pushToken, @JsonKey(name: 'app_version') this.appVersion});
  factory _DeviceCreate.fromJson(Map<String, dynamic> json) => _$DeviceCreateFromJson(json);

@override final  Uuid id;
@override final  DeviceCreatePlatform platform;
@override@JsonKey(name: 'push_token') final  String pushToken;
@override@JsonKey(name: 'app_version') final  String? appVersion;

/// Create a copy of DeviceCreate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DeviceCreateCopyWith<_DeviceCreate> get copyWith => __$DeviceCreateCopyWithImpl<_DeviceCreate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DeviceCreateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _DeviceCreate&&(identical(other.id, id) || other.id == id)&&(identical(other.platform, platform) || other.platform == platform)&&(identical(other.pushToken, pushToken) || other.pushToken == pushToken)&&(identical(other.appVersion, appVersion) || other.appVersion == appVersion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,platform,pushToken,appVersion);
}

@override
String toString() {
    return 'DeviceCreate(id: $id, platform: $platform, pushToken: $pushToken, appVersion: $appVersion)';
}


}

/// @nodoc
abstract mixin class _$DeviceCreateCopyWith<$Res> implements $DeviceCreateCopyWith<$Res> {
  factory _$DeviceCreateCopyWith(_DeviceCreate value, $Res Function(_DeviceCreate) _then) = __$DeviceCreateCopyWithImpl;
@override @useResult
$Res call({
 Uuid id, DeviceCreatePlatform platform,@JsonKey(name: 'push_token') String pushToken,@JsonKey(name: 'app_version') String? appVersion
});




}
/// @nodoc
class __$DeviceCreateCopyWithImpl<$Res>
    implements _$DeviceCreateCopyWith<$Res> {
  __$DeviceCreateCopyWithImpl(this._self, this._then);

  final _DeviceCreate _self;
  final $Res Function(_DeviceCreate) _then;

/// Create a copy of DeviceCreate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? platform = null,Object? pushToken = null,Object? appVersion = freezed,}) {
  return _then(_DeviceCreate(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,platform: null == platform ? _self.platform : platform // ignore: cast_nullable_to_non_nullable
as DeviceCreatePlatform,pushToken: null == pushToken ? _self.pushToken : pushToken // ignore: cast_nullable_to_non_nullable
as String,appVersion: freezed == appVersion ? _self.appVersion : appVersion // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}


}

// dart format on
