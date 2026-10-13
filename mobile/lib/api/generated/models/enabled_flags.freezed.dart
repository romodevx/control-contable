// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'enabled_flags.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$EnabledFlags {

 List<ModuleKey> get enabled;
/// Create a copy of EnabledFlags
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$EnabledFlagsCopyWith<EnabledFlags> get copyWith => _$EnabledFlagsCopyWithImpl<EnabledFlags>(this as EnabledFlags, _$identity);

  /// Serializes this EnabledFlags to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as EnabledFlags;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is EnabledFlags&&const DeepCollectionEquality().equals(other.enabled, _this.enabled));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as EnabledFlags;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.enabled));
}

@override
String toString() {
  final _this = this as EnabledFlags;
  return 'EnabledFlags(enabled: ${_this.enabled})';
}


}

/// @nodoc
abstract mixin class $EnabledFlagsCopyWith<$Res>  {
  factory $EnabledFlagsCopyWith(EnabledFlags value, $Res Function(EnabledFlags) _then) = _$EnabledFlagsCopyWithImpl;
@useResult
$Res call({
 List<ModuleKey> enabled
});




}
/// @nodoc
class _$EnabledFlagsCopyWithImpl<$Res>
    implements $EnabledFlagsCopyWith<$Res> {
  _$EnabledFlagsCopyWithImpl(this._self, this._then);

  final EnabledFlags _self;
  final $Res Function(EnabledFlags) _then;

/// Create a copy of EnabledFlags
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? enabled = null,}) {
  return _then(EnabledFlags(
enabled: null == enabled ? _self.enabled : enabled // ignore: cast_nullable_to_non_nullable
as List<ModuleKey>,
  ));
}

}


/// Adds pattern-matching-related methods to [EnabledFlags].
extension EnabledFlagsPatterns on EnabledFlags {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _EnabledFlags value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _EnabledFlags() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _EnabledFlags value)  $default,){
final _that = this;
switch (_that) {
case _EnabledFlags():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _EnabledFlags value)?  $default,){
final _that = this;
switch (_that) {
case _EnabledFlags() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<ModuleKey> enabled)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _EnabledFlags() when $default != null:
return $default(_that.enabled);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<ModuleKey> enabled)  $default,) {final _that = this;
switch (_that) {
case _EnabledFlags():
return $default(_that.enabled);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<ModuleKey> enabled)?  $default,) {final _that = this;
switch (_that) {
case _EnabledFlags() when $default != null:
return $default(_that.enabled);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _EnabledFlags implements EnabledFlags {
  const _EnabledFlags({required  List<ModuleKey> enabled}): _enabled = enabled;
  factory _EnabledFlags.fromJson(Map<String, dynamic> json) => _$EnabledFlagsFromJson(json);

 final  List<ModuleKey> _enabled;
@override List<ModuleKey> get enabled {
  if (_enabled is EqualUnmodifiableListView) return _enabled;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_enabled);
}


/// Create a copy of EnabledFlags
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$EnabledFlagsCopyWith<_EnabledFlags> get copyWith => __$EnabledFlagsCopyWithImpl<_EnabledFlags>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$EnabledFlagsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _EnabledFlags&&const DeepCollectionEquality().equals(other.enabled, _enabled));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_enabled));
}

@override
String toString() {
    return 'EnabledFlags(enabled: $enabled)';
}


}

/// @nodoc
abstract mixin class _$EnabledFlagsCopyWith<$Res> implements $EnabledFlagsCopyWith<$Res> {
  factory _$EnabledFlagsCopyWith(_EnabledFlags value, $Res Function(_EnabledFlags) _then) = __$EnabledFlagsCopyWithImpl;
@override @useResult
$Res call({
 List<ModuleKey> enabled
});




}
/// @nodoc
class __$EnabledFlagsCopyWithImpl<$Res>
    implements _$EnabledFlagsCopyWith<$Res> {
  __$EnabledFlagsCopyWithImpl(this._self, this._then);

  final _EnabledFlags _self;
  final $Res Function(_EnabledFlags) _then;

/// Create a copy of EnabledFlags
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? enabled = null,}) {
  return _then(_EnabledFlags(
enabled: null == enabled ? _self._enabled : enabled // ignore: cast_nullable_to_non_nullable
as List<ModuleKey>,
  ));
}


}

// dart format on
