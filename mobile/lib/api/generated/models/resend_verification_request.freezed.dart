// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'resend_verification_request.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ResendVerificationRequest {

/// Trimmed, lower-cased and NFC-normalised server-side before lookup or hashing.
 String get email;
/// Create a copy of ResendVerificationRequest
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ResendVerificationRequestCopyWith<ResendVerificationRequest> get copyWith => _$ResendVerificationRequestCopyWithImpl<ResendVerificationRequest>(this as ResendVerificationRequest, _$identity);

  /// Serializes this ResendVerificationRequest to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ResendVerificationRequest;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ResendVerificationRequest&&(identical(other.email, _this.email) || other.email == _this.email));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ResendVerificationRequest;
  return Object.hash(runtimeType,_this.email);
}

@override
String toString() {
  final _this = this as ResendVerificationRequest;
  return 'ResendVerificationRequest(email: ${_this.email})';
}


}

/// @nodoc
abstract mixin class $ResendVerificationRequestCopyWith<$Res>  {
  factory $ResendVerificationRequestCopyWith(ResendVerificationRequest value, $Res Function(ResendVerificationRequest) _then) = _$ResendVerificationRequestCopyWithImpl;
@useResult
$Res call({
 String email
});




}
/// @nodoc
class _$ResendVerificationRequestCopyWithImpl<$Res>
    implements $ResendVerificationRequestCopyWith<$Res> {
  _$ResendVerificationRequestCopyWithImpl(this._self, this._then);

  final ResendVerificationRequest _self;
  final $Res Function(ResendVerificationRequest) _then;

/// Create a copy of ResendVerificationRequest
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? email = null,}) {
  return _then(ResendVerificationRequest(
email: null == email ? _self.email : email // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [ResendVerificationRequest].
extension ResendVerificationRequestPatterns on ResendVerificationRequest {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ResendVerificationRequest value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ResendVerificationRequest() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ResendVerificationRequest value)  $default,){
final _that = this;
switch (_that) {
case _ResendVerificationRequest():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ResendVerificationRequest value)?  $default,){
final _that = this;
switch (_that) {
case _ResendVerificationRequest() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String email)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ResendVerificationRequest() when $default != null:
return $default(_that.email);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String email)  $default,) {final _that = this;
switch (_that) {
case _ResendVerificationRequest():
return $default(_that.email);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String email)?  $default,) {final _that = this;
switch (_that) {
case _ResendVerificationRequest() when $default != null:
return $default(_that.email);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ResendVerificationRequest implements ResendVerificationRequest {
  const _ResendVerificationRequest({required this.email});
  factory _ResendVerificationRequest.fromJson(Map<String, dynamic> json) => _$ResendVerificationRequestFromJson(json);

/// Trimmed, lower-cased and NFC-normalised server-side before lookup or hashing.
@override final  String email;

/// Create a copy of ResendVerificationRequest
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ResendVerificationRequestCopyWith<_ResendVerificationRequest> get copyWith => __$ResendVerificationRequestCopyWithImpl<_ResendVerificationRequest>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ResendVerificationRequestToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ResendVerificationRequest&&(identical(other.email, email) || other.email == email));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,email);
}

@override
String toString() {
    return 'ResendVerificationRequest(email: $email)';
}


}

/// @nodoc
abstract mixin class _$ResendVerificationRequestCopyWith<$Res> implements $ResendVerificationRequestCopyWith<$Res> {
  factory _$ResendVerificationRequestCopyWith(_ResendVerificationRequest value, $Res Function(_ResendVerificationRequest) _then) = __$ResendVerificationRequestCopyWithImpl;
@override @useResult
$Res call({
 String email
});




}
/// @nodoc
class __$ResendVerificationRequestCopyWithImpl<$Res>
    implements _$ResendVerificationRequestCopyWith<$Res> {
  __$ResendVerificationRequestCopyWithImpl(this._self, this._then);

  final _ResendVerificationRequest _self;
  final $Res Function(_ResendVerificationRequest) _then;

/// Create a copy of ResendVerificationRequest
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? email = null,}) {
  return _then(_ResendVerificationRequest(
email: null == email ? _self.email : email // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
