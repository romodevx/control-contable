// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'deductibility_hint.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$DeductibilityHint {

@JsonKey(name: 'regime_code') RegimeCode get regimeCode; DeductibilityHintDeductibility get deductibility; String? get note;@JsonKey(name: 'legal_basis') String? get legalBasis;
/// Create a copy of DeductibilityHint
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DeductibilityHintCopyWith<DeductibilityHint> get copyWith => _$DeductibilityHintCopyWithImpl<DeductibilityHint>(this as DeductibilityHint, _$identity);

  /// Serializes this DeductibilityHint to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as DeductibilityHint;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is DeductibilityHint&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.deductibility, _this.deductibility) || other.deductibility == _this.deductibility)&&(identical(other.note, _this.note) || other.note == _this.note)&&(identical(other.legalBasis, _this.legalBasis) || other.legalBasis == _this.legalBasis));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as DeductibilityHint;
  return Object.hash(runtimeType,_this.regimeCode,_this.deductibility,_this.note,_this.legalBasis);
}

@override
String toString() {
  final _this = this as DeductibilityHint;
  return 'DeductibilityHint(regimeCode: ${_this.regimeCode}, deductibility: ${_this.deductibility}, note: ${_this.note}, legalBasis: ${_this.legalBasis})';
}


}

/// @nodoc
abstract mixin class $DeductibilityHintCopyWith<$Res>  {
  factory $DeductibilityHintCopyWith(DeductibilityHint value, $Res Function(DeductibilityHint) _then) = _$DeductibilityHintCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode, DeductibilityHintDeductibility deductibility, String? note,@JsonKey(name: 'legal_basis') String? legalBasis
});




}
/// @nodoc
class _$DeductibilityHintCopyWithImpl<$Res>
    implements $DeductibilityHintCopyWith<$Res> {
  _$DeductibilityHintCopyWithImpl(this._self, this._then);

  final DeductibilityHint _self;
  final $Res Function(DeductibilityHint) _then;

/// Create a copy of DeductibilityHint
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = null,Object? deductibility = null,Object? note = freezed,Object? legalBasis = freezed,}) {
  return _then(DeductibilityHint(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,deductibility: null == deductibility ? _self.deductibility : deductibility // ignore: cast_nullable_to_non_nullable
as DeductibilityHintDeductibility,note: freezed == note ? _self.note : note // ignore: cast_nullable_to_non_nullable
as String?,legalBasis: freezed == legalBasis ? _self.legalBasis : legalBasis // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}

}


/// Adds pattern-matching-related methods to [DeductibilityHint].
extension DeductibilityHintPatterns on DeductibilityHint {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _DeductibilityHint value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _DeductibilityHint() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _DeductibilityHint value)  $default,){
final _that = this;
switch (_that) {
case _DeductibilityHint():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _DeductibilityHint value)?  $default,){
final _that = this;
switch (_that) {
case _DeductibilityHint() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode,  DeductibilityHintDeductibility deductibility,  String? note, @JsonKey(name: 'legal_basis')  String? legalBasis)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _DeductibilityHint() when $default != null:
return $default(_that.regimeCode,_that.deductibility,_that.note,_that.legalBasis);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode,  DeductibilityHintDeductibility deductibility,  String? note, @JsonKey(name: 'legal_basis')  String? legalBasis)  $default,) {final _that = this;
switch (_that) {
case _DeductibilityHint():
return $default(_that.regimeCode,_that.deductibility,_that.note,_that.legalBasis);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode,  DeductibilityHintDeductibility deductibility,  String? note, @JsonKey(name: 'legal_basis')  String? legalBasis)?  $default,) {final _that = this;
switch (_that) {
case _DeductibilityHint() when $default != null:
return $default(_that.regimeCode,_that.deductibility,_that.note,_that.legalBasis);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _DeductibilityHint implements DeductibilityHint {
  const _DeductibilityHint({@JsonKey(name: 'regime_code') required this.regimeCode, required this.deductibility, this.note, @JsonKey(name: 'legal_basis') this.legalBasis});
  factory _DeductibilityHint.fromJson(Map<String, dynamic> json) => _$DeductibilityHintFromJson(json);

@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
@override final  DeductibilityHintDeductibility deductibility;
@override final  String? note;
@override@JsonKey(name: 'legal_basis') final  String? legalBasis;

/// Create a copy of DeductibilityHint
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DeductibilityHintCopyWith<_DeductibilityHint> get copyWith => __$DeductibilityHintCopyWithImpl<_DeductibilityHint>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DeductibilityHintToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _DeductibilityHint&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.deductibility, deductibility) || other.deductibility == deductibility)&&(identical(other.note, note) || other.note == note)&&(identical(other.legalBasis, legalBasis) || other.legalBasis == legalBasis));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,deductibility,note,legalBasis);
}

@override
String toString() {
    return 'DeductibilityHint(regimeCode: $regimeCode, deductibility: $deductibility, note: $note, legalBasis: $legalBasis)';
}


}

/// @nodoc
abstract mixin class _$DeductibilityHintCopyWith<$Res> implements $DeductibilityHintCopyWith<$Res> {
  factory _$DeductibilityHintCopyWith(_DeductibilityHint value, $Res Function(_DeductibilityHint) _then) = __$DeductibilityHintCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode, DeductibilityHintDeductibility deductibility, String? note,@JsonKey(name: 'legal_basis') String? legalBasis
});




}
/// @nodoc
class __$DeductibilityHintCopyWithImpl<$Res>
    implements _$DeductibilityHintCopyWith<$Res> {
  __$DeductibilityHintCopyWithImpl(this._self, this._then);

  final _DeductibilityHint _self;
  final $Res Function(_DeductibilityHint) _then;

/// Create a copy of DeductibilityHint
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = null,Object? deductibility = null,Object? note = freezed,Object? legalBasis = freezed,}) {
  return _then(_DeductibilityHint(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,deductibility: null == deductibility ? _self.deductibility : deductibility // ignore: cast_nullable_to_non_nullable
as DeductibilityHintDeductibility,note: freezed == note ? _self.note : note // ignore: cast_nullable_to_non_nullable
as String?,legalBasis: freezed == legalBasis ? _self.legalBasis : legalBasis // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}


}

// dart format on
