// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'register_accepted.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RegisterAccepted {

 String get status;
/// Create a copy of RegisterAccepted
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RegisterAcceptedCopyWith<RegisterAccepted> get copyWith => _$RegisterAcceptedCopyWithImpl<RegisterAccepted>(this as RegisterAccepted, _$identity);

  /// Serializes this RegisterAccepted to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RegisterAccepted;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RegisterAccepted&&(identical(other.status, _this.status) || other.status == _this.status));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RegisterAccepted;
  return Object.hash(runtimeType,_this.status);
}

@override
String toString() {
  final _this = this as RegisterAccepted;
  return 'RegisterAccepted(status: ${_this.status})';
}


}

/// @nodoc
abstract mixin class $RegisterAcceptedCopyWith<$Res>  {
  factory $RegisterAcceptedCopyWith(RegisterAccepted value, $Res Function(RegisterAccepted) _then) = _$RegisterAcceptedCopyWithImpl;
@useResult
$Res call({
 String status
});




}
/// @nodoc
class _$RegisterAcceptedCopyWithImpl<$Res>
    implements $RegisterAcceptedCopyWith<$Res> {
  _$RegisterAcceptedCopyWithImpl(this._self, this._then);

  final RegisterAccepted _self;
  final $Res Function(RegisterAccepted) _then;

/// Create a copy of RegisterAccepted
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? status = null,}) {
  return _then(RegisterAccepted(
status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [RegisterAccepted].
extension RegisterAcceptedPatterns on RegisterAccepted {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RegisterAccepted value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RegisterAccepted() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RegisterAccepted value)  $default,){
final _that = this;
switch (_that) {
case _RegisterAccepted():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RegisterAccepted value)?  $default,){
final _that = this;
switch (_that) {
case _RegisterAccepted() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String status)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RegisterAccepted() when $default != null:
return $default(_that.status);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String status)  $default,) {final _that = this;
switch (_that) {
case _RegisterAccepted():
return $default(_that.status);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String status)?  $default,) {final _that = this;
switch (_that) {
case _RegisterAccepted() when $default != null:
return $default(_that.status);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RegisterAccepted implements RegisterAccepted {
  const _RegisterAccepted({required this.status});
  factory _RegisterAccepted.fromJson(Map<String, dynamic> json) => _$RegisterAcceptedFromJson(json);

@override final  String status;

/// Create a copy of RegisterAccepted
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RegisterAcceptedCopyWith<_RegisterAccepted> get copyWith => __$RegisterAcceptedCopyWithImpl<_RegisterAccepted>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RegisterAcceptedToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RegisterAccepted&&(identical(other.status, status) || other.status == status));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,status);
}

@override
String toString() {
    return 'RegisterAccepted(status: $status)';
}


}

/// @nodoc
abstract mixin class _$RegisterAcceptedCopyWith<$Res> implements $RegisterAcceptedCopyWith<$Res> {
  factory _$RegisterAcceptedCopyWith(_RegisterAccepted value, $Res Function(_RegisterAccepted) _then) = __$RegisterAcceptedCopyWithImpl;
@override @useResult
$Res call({
 String status
});




}
/// @nodoc
class __$RegisterAcceptedCopyWithImpl<$Res>
    implements _$RegisterAcceptedCopyWith<$Res> {
  __$RegisterAcceptedCopyWithImpl(this._self, this._then);

  final _RegisterAccepted _self;
  final $Res Function(_RegisterAccepted) _then;

/// Create a copy of RegisterAccepted
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? status = null,}) {
  return _then(_RegisterAccepted(
status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
