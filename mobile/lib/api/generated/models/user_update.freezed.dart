// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'user_update.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$UserUpdate {

 String? get name;
/// Create a copy of UserUpdate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$UserUpdateCopyWith<UserUpdate> get copyWith => _$UserUpdateCopyWithImpl<UserUpdate>(this as UserUpdate, _$identity);

  /// Serializes this UserUpdate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as UserUpdate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is UserUpdate&&(identical(other.name, _this.name) || other.name == _this.name));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as UserUpdate;
  return Object.hash(runtimeType,_this.name);
}

@override
String toString() {
  final _this = this as UserUpdate;
  return 'UserUpdate(name: ${_this.name})';
}


}

/// @nodoc
abstract mixin class $UserUpdateCopyWith<$Res>  {
  factory $UserUpdateCopyWith(UserUpdate value, $Res Function(UserUpdate) _then) = _$UserUpdateCopyWithImpl;
@useResult
$Res call({
 String? name
});




}
/// @nodoc
class _$UserUpdateCopyWithImpl<$Res>
    implements $UserUpdateCopyWith<$Res> {
  _$UserUpdateCopyWithImpl(this._self, this._then);

  final UserUpdate _self;
  final $Res Function(UserUpdate) _then;

/// Create a copy of UserUpdate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? name = freezed,}) {
  return _then(UserUpdate(
name: freezed == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}

}


/// Adds pattern-matching-related methods to [UserUpdate].
extension UserUpdatePatterns on UserUpdate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _UserUpdate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _UserUpdate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _UserUpdate value)  $default,){
final _that = this;
switch (_that) {
case _UserUpdate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _UserUpdate value)?  $default,){
final _that = this;
switch (_that) {
case _UserUpdate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String? name)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _UserUpdate() when $default != null:
return $default(_that.name);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String? name)  $default,) {final _that = this;
switch (_that) {
case _UserUpdate():
return $default(_that.name);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String? name)?  $default,) {final _that = this;
switch (_that) {
case _UserUpdate() when $default != null:
return $default(_that.name);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _UserUpdate implements UserUpdate {
  const _UserUpdate({this.name});
  factory _UserUpdate.fromJson(Map<String, dynamic> json) => _$UserUpdateFromJson(json);

@override final  String? name;

/// Create a copy of UserUpdate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$UserUpdateCopyWith<_UserUpdate> get copyWith => __$UserUpdateCopyWithImpl<_UserUpdate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$UserUpdateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _UserUpdate&&(identical(other.name, name) || other.name == name));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,name);
}

@override
String toString() {
    return 'UserUpdate(name: $name)';
}


}

/// @nodoc
abstract mixin class _$UserUpdateCopyWith<$Res> implements $UserUpdateCopyWith<$Res> {
  factory _$UserUpdateCopyWith(_UserUpdate value, $Res Function(_UserUpdate) _then) = __$UserUpdateCopyWithImpl;
@override @useResult
$Res call({
 String? name
});




}
/// @nodoc
class __$UserUpdateCopyWithImpl<$Res>
    implements _$UserUpdateCopyWith<$Res> {
  __$UserUpdateCopyWithImpl(this._self, this._then);

  final _UserUpdate _self;
  final $Res Function(_UserUpdate) _then;

/// Create a copy of UserUpdate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? name = freezed,}) {
  return _then(_UserUpdate(
name: freezed == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}


}

// dart format on
