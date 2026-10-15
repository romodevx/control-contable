// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'salary_statement_update.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$SalaryStatementUpdate {

@JsonKey(includeIfNull: false, name: 'fiscal_year') FiscalYear? get fiscalYear;@JsonKey(includeIfNull: false, name: 'employer_label') EmployerLabel? get employerLabel;@JsonKey(includeIfNull: false, name: 'accumulable_cents') NonNegativeCents? get accumulableCents;@JsonKey(includeIfNull: false, name: 'exempt_cents') NonNegativeCents? get exemptCents;@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? get isrWithheldCents;@JsonKey(includeIfNull: false) bool? get simultaneous;@JsonKey(includeIfNull: false, name: 'left_before_year_end') bool? get leftBeforeYearEnd;@JsonKey(includeIfNull: false, name: 'payer_not_withholding') bool? get payerNotWithholding;
/// Create a copy of SalaryStatementUpdate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$SalaryStatementUpdateCopyWith<SalaryStatementUpdate> get copyWith => _$SalaryStatementUpdateCopyWithImpl<SalaryStatementUpdate>(this as SalaryStatementUpdate, _$identity);

  /// Serializes this SalaryStatementUpdate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as SalaryStatementUpdate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is SalaryStatementUpdate&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&(identical(other.employerLabel, _this.employerLabel) || other.employerLabel == _this.employerLabel)&&(identical(other.accumulableCents, _this.accumulableCents) || other.accumulableCents == _this.accumulableCents)&&(identical(other.exemptCents, _this.exemptCents) || other.exemptCents == _this.exemptCents)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.simultaneous, _this.simultaneous) || other.simultaneous == _this.simultaneous)&&(identical(other.leftBeforeYearEnd, _this.leftBeforeYearEnd) || other.leftBeforeYearEnd == _this.leftBeforeYearEnd)&&(identical(other.payerNotWithholding, _this.payerNotWithholding) || other.payerNotWithholding == _this.payerNotWithholding));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as SalaryStatementUpdate;
  return Object.hash(runtimeType,_this.fiscalYear,_this.employerLabel,_this.accumulableCents,_this.exemptCents,_this.isrWithheldCents,_this.simultaneous,_this.leftBeforeYearEnd,_this.payerNotWithholding);
}

@override
String toString() {
  final _this = this as SalaryStatementUpdate;
  return 'SalaryStatementUpdate(fiscalYear: ${_this.fiscalYear}, employerLabel: ${_this.employerLabel}, accumulableCents: ${_this.accumulableCents}, exemptCents: ${_this.exemptCents}, isrWithheldCents: ${_this.isrWithheldCents}, simultaneous: ${_this.simultaneous}, leftBeforeYearEnd: ${_this.leftBeforeYearEnd}, payerNotWithholding: ${_this.payerNotWithholding})';
}


}

/// @nodoc
abstract mixin class $SalaryStatementUpdateCopyWith<$Res>  {
  factory $SalaryStatementUpdateCopyWith(SalaryStatementUpdate value, $Res Function(SalaryStatementUpdate) _then) = _$SalaryStatementUpdateCopyWithImpl;
@useResult
$Res call({
@JsonKey(includeIfNull: false, name: 'fiscal_year') FiscalYear? fiscalYear,@JsonKey(includeIfNull: false, name: 'employer_label') EmployerLabel? employerLabel,@JsonKey(includeIfNull: false, name: 'accumulable_cents') NonNegativeCents? accumulableCents,@JsonKey(includeIfNull: false, name: 'exempt_cents') NonNegativeCents? exemptCents,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false) bool? simultaneous,@JsonKey(includeIfNull: false, name: 'left_before_year_end') bool? leftBeforeYearEnd,@JsonKey(includeIfNull: false, name: 'payer_not_withholding') bool? payerNotWithholding
});




}
/// @nodoc
class _$SalaryStatementUpdateCopyWithImpl<$Res>
    implements $SalaryStatementUpdateCopyWith<$Res> {
  _$SalaryStatementUpdateCopyWithImpl(this._self, this._then);

  final SalaryStatementUpdate _self;
  final $Res Function(SalaryStatementUpdate) _then;

/// Create a copy of SalaryStatementUpdate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? fiscalYear = freezed,Object? employerLabel = freezed,Object? accumulableCents = freezed,Object? exemptCents = freezed,Object? isrWithheldCents = freezed,Object? simultaneous = freezed,Object? leftBeforeYearEnd = freezed,Object? payerNotWithholding = freezed,}) {
  return _then(SalaryStatementUpdate(
fiscalYear: freezed == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear?,employerLabel: freezed == employerLabel ? _self.employerLabel : employerLabel // ignore: cast_nullable_to_non_nullable
as EmployerLabel?,accumulableCents: freezed == accumulableCents ? _self.accumulableCents : accumulableCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,exemptCents: freezed == exemptCents ? _self.exemptCents : exemptCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,simultaneous: freezed == simultaneous ? _self.simultaneous : simultaneous // ignore: cast_nullable_to_non_nullable
as bool?,leftBeforeYearEnd: freezed == leftBeforeYearEnd ? _self.leftBeforeYearEnd : leftBeforeYearEnd // ignore: cast_nullable_to_non_nullable
as bool?,payerNotWithholding: freezed == payerNotWithholding ? _self.payerNotWithholding : payerNotWithholding // ignore: cast_nullable_to_non_nullable
as bool?,
  ));
}

}


/// Adds pattern-matching-related methods to [SalaryStatementUpdate].
extension SalaryStatementUpdatePatterns on SalaryStatementUpdate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _SalaryStatementUpdate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _SalaryStatementUpdate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _SalaryStatementUpdate value)  $default,){
final _that = this;
switch (_that) {
case _SalaryStatementUpdate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _SalaryStatementUpdate value)?  $default,){
final _that = this;
switch (_that) {
case _SalaryStatementUpdate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: false, name: 'fiscal_year')  FiscalYear? fiscalYear, @JsonKey(includeIfNull: false, name: 'employer_label')  EmployerLabel? employerLabel, @JsonKey(includeIfNull: false, name: 'accumulable_cents')  NonNegativeCents? accumulableCents, @JsonKey(includeIfNull: false, name: 'exempt_cents')  NonNegativeCents? exemptCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false)  bool? simultaneous, @JsonKey(includeIfNull: false, name: 'left_before_year_end')  bool? leftBeforeYearEnd, @JsonKey(includeIfNull: false, name: 'payer_not_withholding')  bool? payerNotWithholding)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _SalaryStatementUpdate() when $default != null:
return $default(_that.fiscalYear,_that.employerLabel,_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents,_that.simultaneous,_that.leftBeforeYearEnd,_that.payerNotWithholding);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: false, name: 'fiscal_year')  FiscalYear? fiscalYear, @JsonKey(includeIfNull: false, name: 'employer_label')  EmployerLabel? employerLabel, @JsonKey(includeIfNull: false, name: 'accumulable_cents')  NonNegativeCents? accumulableCents, @JsonKey(includeIfNull: false, name: 'exempt_cents')  NonNegativeCents? exemptCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false)  bool? simultaneous, @JsonKey(includeIfNull: false, name: 'left_before_year_end')  bool? leftBeforeYearEnd, @JsonKey(includeIfNull: false, name: 'payer_not_withholding')  bool? payerNotWithholding)  $default,) {final _that = this;
switch (_that) {
case _SalaryStatementUpdate():
return $default(_that.fiscalYear,_that.employerLabel,_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents,_that.simultaneous,_that.leftBeforeYearEnd,_that.payerNotWithholding);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(includeIfNull: false, name: 'fiscal_year')  FiscalYear? fiscalYear, @JsonKey(includeIfNull: false, name: 'employer_label')  EmployerLabel? employerLabel, @JsonKey(includeIfNull: false, name: 'accumulable_cents')  NonNegativeCents? accumulableCents, @JsonKey(includeIfNull: false, name: 'exempt_cents')  NonNegativeCents? exemptCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false)  bool? simultaneous, @JsonKey(includeIfNull: false, name: 'left_before_year_end')  bool? leftBeforeYearEnd, @JsonKey(includeIfNull: false, name: 'payer_not_withholding')  bool? payerNotWithholding)?  $default,) {final _that = this;
switch (_that) {
case _SalaryStatementUpdate() when $default != null:
return $default(_that.fiscalYear,_that.employerLabel,_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents,_that.simultaneous,_that.leftBeforeYearEnd,_that.payerNotWithholding);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _SalaryStatementUpdate implements SalaryStatementUpdate {
  const _SalaryStatementUpdate({@JsonKey(includeIfNull: false, name: 'fiscal_year') this.fiscalYear, @JsonKey(includeIfNull: false, name: 'employer_label') this.employerLabel, @JsonKey(includeIfNull: false, name: 'accumulable_cents') this.accumulableCents, @JsonKey(includeIfNull: false, name: 'exempt_cents') this.exemptCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents') this.isrWithheldCents, @JsonKey(includeIfNull: false) this.simultaneous, @JsonKey(includeIfNull: false, name: 'left_before_year_end') this.leftBeforeYearEnd, @JsonKey(includeIfNull: false, name: 'payer_not_withholding') this.payerNotWithholding});
  factory _SalaryStatementUpdate.fromJson(Map<String, dynamic> json) => _$SalaryStatementUpdateFromJson(json);

@override@JsonKey(includeIfNull: false, name: 'fiscal_year') final  FiscalYear? fiscalYear;
@override@JsonKey(includeIfNull: false, name: 'employer_label') final  EmployerLabel? employerLabel;
@override@JsonKey(includeIfNull: false, name: 'accumulable_cents') final  NonNegativeCents? accumulableCents;
@override@JsonKey(includeIfNull: false, name: 'exempt_cents') final  NonNegativeCents? exemptCents;
@override@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') final  NonNegativeCents? isrWithheldCents;
@override@JsonKey(includeIfNull: false) final  bool? simultaneous;
@override@JsonKey(includeIfNull: false, name: 'left_before_year_end') final  bool? leftBeforeYearEnd;
@override@JsonKey(includeIfNull: false, name: 'payer_not_withholding') final  bool? payerNotWithholding;

/// Create a copy of SalaryStatementUpdate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$SalaryStatementUpdateCopyWith<_SalaryStatementUpdate> get copyWith => __$SalaryStatementUpdateCopyWithImpl<_SalaryStatementUpdate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$SalaryStatementUpdateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _SalaryStatementUpdate&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&(identical(other.employerLabel, employerLabel) || other.employerLabel == employerLabel)&&(identical(other.accumulableCents, accumulableCents) || other.accumulableCents == accumulableCents)&&(identical(other.exemptCents, exemptCents) || other.exemptCents == exemptCents)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.simultaneous, simultaneous) || other.simultaneous == simultaneous)&&(identical(other.leftBeforeYearEnd, leftBeforeYearEnd) || other.leftBeforeYearEnd == leftBeforeYearEnd)&&(identical(other.payerNotWithholding, payerNotWithholding) || other.payerNotWithholding == payerNotWithholding));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,fiscalYear,employerLabel,accumulableCents,exemptCents,isrWithheldCents,simultaneous,leftBeforeYearEnd,payerNotWithholding);
}

@override
String toString() {
    return 'SalaryStatementUpdate(fiscalYear: $fiscalYear, employerLabel: $employerLabel, accumulableCents: $accumulableCents, exemptCents: $exemptCents, isrWithheldCents: $isrWithheldCents, simultaneous: $simultaneous, leftBeforeYearEnd: $leftBeforeYearEnd, payerNotWithholding: $payerNotWithholding)';
}


}

/// @nodoc
abstract mixin class _$SalaryStatementUpdateCopyWith<$Res> implements $SalaryStatementUpdateCopyWith<$Res> {
  factory _$SalaryStatementUpdateCopyWith(_SalaryStatementUpdate value, $Res Function(_SalaryStatementUpdate) _then) = __$SalaryStatementUpdateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(includeIfNull: false, name: 'fiscal_year') FiscalYear? fiscalYear,@JsonKey(includeIfNull: false, name: 'employer_label') EmployerLabel? employerLabel,@JsonKey(includeIfNull: false, name: 'accumulable_cents') NonNegativeCents? accumulableCents,@JsonKey(includeIfNull: false, name: 'exempt_cents') NonNegativeCents? exemptCents,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false) bool? simultaneous,@JsonKey(includeIfNull: false, name: 'left_before_year_end') bool? leftBeforeYearEnd,@JsonKey(includeIfNull: false, name: 'payer_not_withholding') bool? payerNotWithholding
});




}
/// @nodoc
class __$SalaryStatementUpdateCopyWithImpl<$Res>
    implements _$SalaryStatementUpdateCopyWith<$Res> {
  __$SalaryStatementUpdateCopyWithImpl(this._self, this._then);

  final _SalaryStatementUpdate _self;
  final $Res Function(_SalaryStatementUpdate) _then;

/// Create a copy of SalaryStatementUpdate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? fiscalYear = freezed,Object? employerLabel = freezed,Object? accumulableCents = freezed,Object? exemptCents = freezed,Object? isrWithheldCents = freezed,Object? simultaneous = freezed,Object? leftBeforeYearEnd = freezed,Object? payerNotWithholding = freezed,}) {
  return _then(_SalaryStatementUpdate(
fiscalYear: freezed == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear?,employerLabel: freezed == employerLabel ? _self.employerLabel : employerLabel // ignore: cast_nullable_to_non_nullable
as EmployerLabel?,accumulableCents: freezed == accumulableCents ? _self.accumulableCents : accumulableCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,exemptCents: freezed == exemptCents ? _self.exemptCents : exemptCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,simultaneous: freezed == simultaneous ? _self.simultaneous : simultaneous // ignore: cast_nullable_to_non_nullable
as bool?,leftBeforeYearEnd: freezed == leftBeforeYearEnd ? _self.leftBeforeYearEnd : leftBeforeYearEnd // ignore: cast_nullable_to_non_nullable
as bool?,payerNotWithholding: freezed == payerNotWithholding ? _self.payerNotWithholding : payerNotWithholding // ignore: cast_nullable_to_non_nullable
as bool?,
  ));
}


}

// dart format on
