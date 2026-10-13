// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'verification_sent.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$VerificationSent {

 String get status;
/// Create a copy of VerificationSent
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$VerificationSentCopyWith<VerificationSent> get copyWith => _$VerificationSentCopyWithImpl<VerificationSent>(this as VerificationSent, _$identity);

  /// Serializes this VerificationSent to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as VerificationSent;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is VerificationSent&&(identical(other.status, _this.status) || other.status == _this.status));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as VerificationSent;
  return Object.hash(runtimeType,_this.status);
}

@override
String toString() {
  final _this = this as VerificationSent;
  return 'VerificationSent(status: ${_this.status})';
}


}

/// @nodoc
abstract mixin class $VerificationSentCopyWith<$Res>  {
  factory $VerificationSentCopyWith(VerificationSent value, $Res Function(VerificationSent) _then) = _$VerificationSentCopyWithImpl;
@useResult
$Res call({
 String status
});




}
/// @nodoc
class _$VerificationSentCopyWithImpl<$Res>
    implements $VerificationSentCopyWith<$Res> {
  _$VerificationSentCopyWithImpl(this._self, this._then);

  final VerificationSent _self;
  final $Res Function(VerificationSent) _then;

/// Create a copy of VerificationSent
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? status = null,}) {
  return _then(VerificationSent(
status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [VerificationSent].
extension VerificationSentPatterns on VerificationSent {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _VerificationSent value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _VerificationSent() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _VerificationSent value)  $default,){
final _that = this;
switch (_that) {
case _VerificationSent():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _VerificationSent value)?  $default,){
final _that = this;
switch (_that) {
case _VerificationSent() when $default != null:
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
case _VerificationSent() when $default != null:
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
case _VerificationSent():
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
case _VerificationSent() when $default != null:
return $default(_that.status);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _VerificationSent implements VerificationSent {
  const _VerificationSent({required this.status});
  factory _VerificationSent.fromJson(Map<String, dynamic> json) => _$VerificationSentFromJson(json);

@override final  String status;

/// Create a copy of VerificationSent
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$VerificationSentCopyWith<_VerificationSent> get copyWith => __$VerificationSentCopyWithImpl<_VerificationSent>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$VerificationSentToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _VerificationSent&&(identical(other.status, status) || other.status == status));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,status);
}

@override
String toString() {
    return 'VerificationSent(status: $status)';
}


}

/// @nodoc
abstract mixin class _$VerificationSentCopyWith<$Res> implements $VerificationSentCopyWith<$Res> {
  factory _$VerificationSentCopyWith(_VerificationSent value, $Res Function(_VerificationSent) _then) = __$VerificationSentCopyWithImpl;
@override @useResult
$Res call({
 String status
});




}
/// @nodoc
class __$VerificationSentCopyWithImpl<$Res>
    implements _$VerificationSentCopyWith<$Res> {
  __$VerificationSentCopyWithImpl(this._self, this._then);

  final _VerificationSent _self;
  final $Res Function(_VerificationSent) _then;

/// Create a copy of VerificationSent
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? status = null,}) {
  return _then(_VerificationSent(
status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
