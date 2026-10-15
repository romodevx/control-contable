// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'filed_payments.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$FiledPayments {

@JsonKey(name: 'regime_code') RegimeCode get regimeCode;@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear; List<FiledPaymentMonth> get months;@JsonKey(name: 'total_effective_cents') NonNegativeCents get totalEffectiveCents;@JsonKey(includeIfNull: true, name: 'updated_at') DateTime? get updatedAt;
/// Create a copy of FiledPayments
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$FiledPaymentsCopyWith<FiledPayments> get copyWith => _$FiledPaymentsCopyWithImpl<FiledPayments>(this as FiledPayments, _$identity);

  /// Serializes this FiledPayments to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as FiledPayments;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is FiledPayments&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&const DeepCollectionEquality().equals(other.months, _this.months)&&(identical(other.totalEffectiveCents, _this.totalEffectiveCents) || other.totalEffectiveCents == _this.totalEffectiveCents)&&(identical(other.updatedAt, _this.updatedAt) || other.updatedAt == _this.updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as FiledPayments;
  return Object.hash(runtimeType,_this.regimeCode,_this.fiscalYear,const DeepCollectionEquality().hash(_this.months),_this.totalEffectiveCents,_this.updatedAt);
}

@override
String toString() {
  final _this = this as FiledPayments;
  return 'FiledPayments(regimeCode: ${_this.regimeCode}, fiscalYear: ${_this.fiscalYear}, months: ${_this.months}, totalEffectiveCents: ${_this.totalEffectiveCents}, updatedAt: ${_this.updatedAt})';
}


}

/// @nodoc
abstract mixin class $FiledPaymentsCopyWith<$Res>  {
  factory $FiledPaymentsCopyWith(FiledPayments value, $Res Function(FiledPayments) _then) = _$FiledPaymentsCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, List<FiledPaymentMonth> months,@JsonKey(name: 'total_effective_cents') NonNegativeCents totalEffectiveCents,@JsonKey(includeIfNull: true, name: 'updated_at') DateTime? updatedAt
});




}
/// @nodoc
class _$FiledPaymentsCopyWithImpl<$Res>
    implements $FiledPaymentsCopyWith<$Res> {
  _$FiledPaymentsCopyWithImpl(this._self, this._then);

  final FiledPayments _self;
  final $Res Function(FiledPayments) _then;

/// Create a copy of FiledPayments
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = null,Object? fiscalYear = null,Object? months = null,Object? totalEffectiveCents = null,Object? updatedAt = freezed,}) {
  return _then(FiledPayments(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,months: null == months ? _self.months : months // ignore: cast_nullable_to_non_nullable
as List<FiledPaymentMonth>,totalEffectiveCents: null == totalEffectiveCents ? _self.totalEffectiveCents : totalEffectiveCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}

}


/// Adds pattern-matching-related methods to [FiledPayments].
extension FiledPaymentsPatterns on FiledPayments {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _FiledPayments value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _FiledPayments() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _FiledPayments value)  $default,){
final _that = this;
switch (_that) {
case _FiledPayments():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _FiledPayments value)?  $default,){
final _that = this;
switch (_that) {
case _FiledPayments() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<FiledPaymentMonth> months, @JsonKey(name: 'total_effective_cents')  NonNegativeCents totalEffectiveCents, @JsonKey(includeIfNull: true, name: 'updated_at')  DateTime? updatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _FiledPayments() when $default != null:
return $default(_that.regimeCode,_that.fiscalYear,_that.months,_that.totalEffectiveCents,_that.updatedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<FiledPaymentMonth> months, @JsonKey(name: 'total_effective_cents')  NonNegativeCents totalEffectiveCents, @JsonKey(includeIfNull: true, name: 'updated_at')  DateTime? updatedAt)  $default,) {final _that = this;
switch (_that) {
case _FiledPayments():
return $default(_that.regimeCode,_that.fiscalYear,_that.months,_that.totalEffectiveCents,_that.updatedAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<FiledPaymentMonth> months, @JsonKey(name: 'total_effective_cents')  NonNegativeCents totalEffectiveCents, @JsonKey(includeIfNull: true, name: 'updated_at')  DateTime? updatedAt)?  $default,) {final _that = this;
switch (_that) {
case _FiledPayments() when $default != null:
return $default(_that.regimeCode,_that.fiscalYear,_that.months,_that.totalEffectiveCents,_that.updatedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _FiledPayments implements FiledPayments {
  const _FiledPayments({@JsonKey(name: 'regime_code') required this.regimeCode, @JsonKey(name: 'fiscal_year') required this.fiscalYear, required  List<FiledPaymentMonth> months, @JsonKey(name: 'total_effective_cents') required this.totalEffectiveCents, @JsonKey(includeIfNull: true, name: 'updated_at') required this.updatedAt}): _months = months;
  factory _FiledPayments.fromJson(Map<String, dynamic> json) => _$FiledPaymentsFromJson(json);

@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
 final  List<FiledPaymentMonth> _months;
@override List<FiledPaymentMonth> get months {
  if (_months is EqualUnmodifiableListView) return _months;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_months);
}

@override@JsonKey(name: 'total_effective_cents') final  NonNegativeCents totalEffectiveCents;
@override@JsonKey(includeIfNull: true, name: 'updated_at') final  DateTime? updatedAt;

/// Create a copy of FiledPayments
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$FiledPaymentsCopyWith<_FiledPayments> get copyWith => __$FiledPaymentsCopyWithImpl<_FiledPayments>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$FiledPaymentsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _FiledPayments&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&const DeepCollectionEquality().equals(other.months, _months)&&(identical(other.totalEffectiveCents, totalEffectiveCents) || other.totalEffectiveCents == totalEffectiveCents)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,fiscalYear,const DeepCollectionEquality().hash(_months),totalEffectiveCents,updatedAt);
}

@override
String toString() {
    return 'FiledPayments(regimeCode: $regimeCode, fiscalYear: $fiscalYear, months: $months, totalEffectiveCents: $totalEffectiveCents, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class _$FiledPaymentsCopyWith<$Res> implements $FiledPaymentsCopyWith<$Res> {
  factory _$FiledPaymentsCopyWith(_FiledPayments value, $Res Function(_FiledPayments) _then) = __$FiledPaymentsCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, List<FiledPaymentMonth> months,@JsonKey(name: 'total_effective_cents') NonNegativeCents totalEffectiveCents,@JsonKey(includeIfNull: true, name: 'updated_at') DateTime? updatedAt
});




}
/// @nodoc
class __$FiledPaymentsCopyWithImpl<$Res>
    implements _$FiledPaymentsCopyWith<$Res> {
  __$FiledPaymentsCopyWithImpl(this._self, this._then);

  final _FiledPayments _self;
  final $Res Function(_FiledPayments) _then;

/// Create a copy of FiledPayments
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = null,Object? fiscalYear = null,Object? months = null,Object? totalEffectiveCents = null,Object? updatedAt = freezed,}) {
  return _then(_FiledPayments(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,months: null == months ? _self._months : months // ignore: cast_nullable_to_non_nullable
as List<FiledPaymentMonth>,totalEffectiveCents: null == totalEffectiveCents ? _self.totalEffectiveCents : totalEffectiveCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,updatedAt: freezed == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}


}

// dart format on
