// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'regime_incompatibility.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RegimeIncompatibility {

 RegimeCode get code;/// Plain-language reason shown to the user.
 String get reason;
/// Create a copy of RegimeIncompatibility
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RegimeIncompatibilityCopyWith<RegimeIncompatibility> get copyWith => _$RegimeIncompatibilityCopyWithImpl<RegimeIncompatibility>(this as RegimeIncompatibility, _$identity);

  /// Serializes this RegimeIncompatibility to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RegimeIncompatibility;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RegimeIncompatibility&&(identical(other.code, _this.code) || other.code == _this.code)&&(identical(other.reason, _this.reason) || other.reason == _this.reason));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RegimeIncompatibility;
  return Object.hash(runtimeType,_this.code,_this.reason);
}

@override
String toString() {
  final _this = this as RegimeIncompatibility;
  return 'RegimeIncompatibility(code: ${_this.code}, reason: ${_this.reason})';
}


}

/// @nodoc
abstract mixin class $RegimeIncompatibilityCopyWith<$Res>  {
  factory $RegimeIncompatibilityCopyWith(RegimeIncompatibility value, $Res Function(RegimeIncompatibility) _then) = _$RegimeIncompatibilityCopyWithImpl;
@useResult
$Res call({
 RegimeCode code, String reason
});




}
/// @nodoc
class _$RegimeIncompatibilityCopyWithImpl<$Res>
    implements $RegimeIncompatibilityCopyWith<$Res> {
  _$RegimeIncompatibilityCopyWithImpl(this._self, this._then);

  final RegimeIncompatibility _self;
  final $Res Function(RegimeIncompatibility) _then;

/// Create a copy of RegimeIncompatibility
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? code = null,Object? reason = null,}) {
  return _then(RegimeIncompatibility(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as RegimeCode,reason: null == reason ? _self.reason : reason // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [RegimeIncompatibility].
extension RegimeIncompatibilityPatterns on RegimeIncompatibility {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RegimeIncompatibility value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RegimeIncompatibility() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RegimeIncompatibility value)  $default,){
final _that = this;
switch (_that) {
case _RegimeIncompatibility():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RegimeIncompatibility value)?  $default,){
final _that = this;
switch (_that) {
case _RegimeIncompatibility() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( RegimeCode code,  String reason)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RegimeIncompatibility() when $default != null:
return $default(_that.code,_that.reason);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( RegimeCode code,  String reason)  $default,) {final _that = this;
switch (_that) {
case _RegimeIncompatibility():
return $default(_that.code,_that.reason);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( RegimeCode code,  String reason)?  $default,) {final _that = this;
switch (_that) {
case _RegimeIncompatibility() when $default != null:
return $default(_that.code,_that.reason);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RegimeIncompatibility implements RegimeIncompatibility {
  const _RegimeIncompatibility({required this.code, required this.reason});
  factory _RegimeIncompatibility.fromJson(Map<String, dynamic> json) => _$RegimeIncompatibilityFromJson(json);

@override final  RegimeCode code;
/// Plain-language reason shown to the user.
@override final  String reason;

/// Create a copy of RegimeIncompatibility
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RegimeIncompatibilityCopyWith<_RegimeIncompatibility> get copyWith => __$RegimeIncompatibilityCopyWithImpl<_RegimeIncompatibility>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RegimeIncompatibilityToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RegimeIncompatibility&&(identical(other.code, code) || other.code == code)&&(identical(other.reason, reason) || other.reason == reason));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,code,reason);
}

@override
String toString() {
    return 'RegimeIncompatibility(code: $code, reason: $reason)';
}


}

/// @nodoc
abstract mixin class _$RegimeIncompatibilityCopyWith<$Res> implements $RegimeIncompatibilityCopyWith<$Res> {
  factory _$RegimeIncompatibilityCopyWith(_RegimeIncompatibility value, $Res Function(_RegimeIncompatibility) _then) = __$RegimeIncompatibilityCopyWithImpl;
@override @useResult
$Res call({
 RegimeCode code, String reason
});




}
/// @nodoc
class __$RegimeIncompatibilityCopyWithImpl<$Res>
    implements _$RegimeIncompatibilityCopyWith<$Res> {
  __$RegimeIncompatibilityCopyWithImpl(this._self, this._then);

  final _RegimeIncompatibility _self;
  final $Res Function(_RegimeIncompatibility) _then;

/// Create a copy of RegimeIncompatibility
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? code = null,Object? reason = null,}) {
  return _then(_RegimeIncompatibility(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as RegimeCode,reason: null == reason ? _self.reason : reason // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
