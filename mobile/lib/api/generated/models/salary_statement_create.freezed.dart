// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'salary_statement_create.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$SalaryStatementCreate {

@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear;@JsonKey(name: 'accumulable_cents') NonNegativeCents get accumulableCents;@JsonKey(name: 'exempt_cents') NonNegativeCents get exemptCents;@JsonKey(name: 'isr_withheld_cents') NonNegativeCents get isrWithheldCents; bool get simultaneous;@JsonKey(name: 'left_before_year_end') bool get leftBeforeYearEnd;@JsonKey(name: 'payer_not_withholding') bool get payerNotWithholding;@JsonKey(includeIfNull: false) Uuid? get id;@JsonKey(includeIfNull: false, name: 'employer_label') EmployerLabel? get employerLabel;
/// Create a copy of SalaryStatementCreate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$SalaryStatementCreateCopyWith<SalaryStatementCreate> get copyWith => _$SalaryStatementCreateCopyWithImpl<SalaryStatementCreate>(this as SalaryStatementCreate, _$identity);

  /// Serializes this SalaryStatementCreate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as SalaryStatementCreate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is SalaryStatementCreate&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&(identical(other.accumulableCents, _this.accumulableCents) || other.accumulableCents == _this.accumulableCents)&&(identical(other.exemptCents, _this.exemptCents) || other.exemptCents == _this.exemptCents)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.simultaneous, _this.simultaneous) || other.simultaneous == _this.simultaneous)&&(identical(other.leftBeforeYearEnd, _this.leftBeforeYearEnd) || other.leftBeforeYearEnd == _this.leftBeforeYearEnd)&&(identical(other.payerNotWithholding, _this.payerNotWithholding) || other.payerNotWithholding == _this.payerNotWithholding)&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.employerLabel, _this.employerLabel) || other.employerLabel == _this.employerLabel));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as SalaryStatementCreate;
  return Object.hash(runtimeType,_this.fiscalYear,_this.accumulableCents,_this.exemptCents,_this.isrWithheldCents,_this.simultaneous,_this.leftBeforeYearEnd,_this.payerNotWithholding,_this.id,_this.employerLabel);
}

@override
String toString() {
  final _this = this as SalaryStatementCreate;
  return 'SalaryStatementCreate(fiscalYear: ${_this.fiscalYear}, accumulableCents: ${_this.accumulableCents}, exemptCents: ${_this.exemptCents}, isrWithheldCents: ${_this.isrWithheldCents}, simultaneous: ${_this.simultaneous}, leftBeforeYearEnd: ${_this.leftBeforeYearEnd}, payerNotWithholding: ${_this.payerNotWithholding}, id: ${_this.id}, employerLabel: ${_this.employerLabel})';
}


}

/// @nodoc
abstract mixin class $SalaryStatementCreateCopyWith<$Res>  {
  factory $SalaryStatementCreateCopyWith(SalaryStatementCreate value, $Res Function(SalaryStatementCreate) _then) = _$SalaryStatementCreateCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear,@JsonKey(name: 'accumulable_cents') NonNegativeCents accumulableCents,@JsonKey(name: 'exempt_cents') NonNegativeCents exemptCents,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents isrWithheldCents, bool simultaneous,@JsonKey(name: 'left_before_year_end') bool leftBeforeYearEnd,@JsonKey(name: 'payer_not_withholding') bool payerNotWithholding,@JsonKey(includeIfNull: false) Uuid? id,@JsonKey(includeIfNull: false, name: 'employer_label') EmployerLabel? employerLabel
});




}
/// @nodoc
class _$SalaryStatementCreateCopyWithImpl<$Res>
    implements $SalaryStatementCreateCopyWith<$Res> {
  _$SalaryStatementCreateCopyWithImpl(this._self, this._then);

  final SalaryStatementCreate _self;
  final $Res Function(SalaryStatementCreate) _then;

/// Create a copy of SalaryStatementCreate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? fiscalYear = null,Object? accumulableCents = null,Object? exemptCents = null,Object? isrWithheldCents = null,Object? simultaneous = null,Object? leftBeforeYearEnd = null,Object? payerNotWithholding = null,Object? id = freezed,Object? employerLabel = freezed,}) {
  return _then(SalaryStatementCreate(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,accumulableCents: null == accumulableCents ? _self.accumulableCents : accumulableCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,exemptCents: null == exemptCents ? _self.exemptCents : exemptCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrWithheldCents: null == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,simultaneous: null == simultaneous ? _self.simultaneous : simultaneous // ignore: cast_nullable_to_non_nullable
as bool,leftBeforeYearEnd: null == leftBeforeYearEnd ? _self.leftBeforeYearEnd : leftBeforeYearEnd // ignore: cast_nullable_to_non_nullable
as bool,payerNotWithholding: null == payerNotWithholding ? _self.payerNotWithholding : payerNotWithholding // ignore: cast_nullable_to_non_nullable
as bool,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,employerLabel: freezed == employerLabel ? _self.employerLabel : employerLabel // ignore: cast_nullable_to_non_nullable
as EmployerLabel?,
  ));
}

}


/// Adds pattern-matching-related methods to [SalaryStatementCreate].
extension SalaryStatementCreatePatterns on SalaryStatementCreate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _SalaryStatementCreate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _SalaryStatementCreate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _SalaryStatementCreate value)  $default,){
final _that = this;
switch (_that) {
case _SalaryStatementCreate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _SalaryStatementCreate value)?  $default,){
final _that = this;
switch (_that) {
case _SalaryStatementCreate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(name: 'accumulable_cents')  NonNegativeCents accumulableCents, @JsonKey(name: 'exempt_cents')  NonNegativeCents exemptCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents,  bool simultaneous, @JsonKey(name: 'left_before_year_end')  bool leftBeforeYearEnd, @JsonKey(name: 'payer_not_withholding')  bool payerNotWithholding, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'employer_label')  EmployerLabel? employerLabel)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _SalaryStatementCreate() when $default != null:
return $default(_that.fiscalYear,_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents,_that.simultaneous,_that.leftBeforeYearEnd,_that.payerNotWithholding,_that.id,_that.employerLabel);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(name: 'accumulable_cents')  NonNegativeCents accumulableCents, @JsonKey(name: 'exempt_cents')  NonNegativeCents exemptCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents,  bool simultaneous, @JsonKey(name: 'left_before_year_end')  bool leftBeforeYearEnd, @JsonKey(name: 'payer_not_withholding')  bool payerNotWithholding, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'employer_label')  EmployerLabel? employerLabel)  $default,) {final _that = this;
switch (_that) {
case _SalaryStatementCreate():
return $default(_that.fiscalYear,_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents,_that.simultaneous,_that.leftBeforeYearEnd,_that.payerNotWithholding,_that.id,_that.employerLabel);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(name: 'accumulable_cents')  NonNegativeCents accumulableCents, @JsonKey(name: 'exempt_cents')  NonNegativeCents exemptCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents,  bool simultaneous, @JsonKey(name: 'left_before_year_end')  bool leftBeforeYearEnd, @JsonKey(name: 'payer_not_withholding')  bool payerNotWithholding, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'employer_label')  EmployerLabel? employerLabel)?  $default,) {final _that = this;
switch (_that) {
case _SalaryStatementCreate() when $default != null:
return $default(_that.fiscalYear,_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents,_that.simultaneous,_that.leftBeforeYearEnd,_that.payerNotWithholding,_that.id,_that.employerLabel);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _SalaryStatementCreate implements SalaryStatementCreate {
  const _SalaryStatementCreate({@JsonKey(name: 'fiscal_year') required this.fiscalYear, @JsonKey(name: 'accumulable_cents') required this.accumulableCents, @JsonKey(name: 'exempt_cents') required this.exemptCents, @JsonKey(name: 'isr_withheld_cents') required this.isrWithheldCents, this.simultaneous = false, @JsonKey(name: 'left_before_year_end') this.leftBeforeYearEnd = false, @JsonKey(name: 'payer_not_withholding') this.payerNotWithholding = false, @JsonKey(includeIfNull: false) this.id, @JsonKey(includeIfNull: false, name: 'employer_label') this.employerLabel});
  factory _SalaryStatementCreate.fromJson(Map<String, dynamic> json) => _$SalaryStatementCreateFromJson(json);

@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
@override@JsonKey(name: 'accumulable_cents') final  NonNegativeCents accumulableCents;
@override@JsonKey(name: 'exempt_cents') final  NonNegativeCents exemptCents;
@override@JsonKey(name: 'isr_withheld_cents') final  NonNegativeCents isrWithheldCents;
@override@JsonKey() final  bool simultaneous;
@override@JsonKey(name: 'left_before_year_end') final  bool leftBeforeYearEnd;
@override@JsonKey(name: 'payer_not_withholding') final  bool payerNotWithholding;
@override@JsonKey(includeIfNull: false) final  Uuid? id;
@override@JsonKey(includeIfNull: false, name: 'employer_label') final  EmployerLabel? employerLabel;

/// Create a copy of SalaryStatementCreate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$SalaryStatementCreateCopyWith<_SalaryStatementCreate> get copyWith => __$SalaryStatementCreateCopyWithImpl<_SalaryStatementCreate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$SalaryStatementCreateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _SalaryStatementCreate&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&(identical(other.accumulableCents, accumulableCents) || other.accumulableCents == accumulableCents)&&(identical(other.exemptCents, exemptCents) || other.exemptCents == exemptCents)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.simultaneous, simultaneous) || other.simultaneous == simultaneous)&&(identical(other.leftBeforeYearEnd, leftBeforeYearEnd) || other.leftBeforeYearEnd == leftBeforeYearEnd)&&(identical(other.payerNotWithholding, payerNotWithholding) || other.payerNotWithholding == payerNotWithholding)&&(identical(other.id, id) || other.id == id)&&(identical(other.employerLabel, employerLabel) || other.employerLabel == employerLabel));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,fiscalYear,accumulableCents,exemptCents,isrWithheldCents,simultaneous,leftBeforeYearEnd,payerNotWithholding,id,employerLabel);
}

@override
String toString() {
    return 'SalaryStatementCreate(fiscalYear: $fiscalYear, accumulableCents: $accumulableCents, exemptCents: $exemptCents, isrWithheldCents: $isrWithheldCents, simultaneous: $simultaneous, leftBeforeYearEnd: $leftBeforeYearEnd, payerNotWithholding: $payerNotWithholding, id: $id, employerLabel: $employerLabel)';
}


}

/// @nodoc
abstract mixin class _$SalaryStatementCreateCopyWith<$Res> implements $SalaryStatementCreateCopyWith<$Res> {
  factory _$SalaryStatementCreateCopyWith(_SalaryStatementCreate value, $Res Function(_SalaryStatementCreate) _then) = __$SalaryStatementCreateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear,@JsonKey(name: 'accumulable_cents') NonNegativeCents accumulableCents,@JsonKey(name: 'exempt_cents') NonNegativeCents exemptCents,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents isrWithheldCents, bool simultaneous,@JsonKey(name: 'left_before_year_end') bool leftBeforeYearEnd,@JsonKey(name: 'payer_not_withholding') bool payerNotWithholding,@JsonKey(includeIfNull: false) Uuid? id,@JsonKey(includeIfNull: false, name: 'employer_label') EmployerLabel? employerLabel
});




}
/// @nodoc
class __$SalaryStatementCreateCopyWithImpl<$Res>
    implements _$SalaryStatementCreateCopyWith<$Res> {
  __$SalaryStatementCreateCopyWithImpl(this._self, this._then);

  final _SalaryStatementCreate _self;
  final $Res Function(_SalaryStatementCreate) _then;

/// Create a copy of SalaryStatementCreate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? fiscalYear = null,Object? accumulableCents = null,Object? exemptCents = null,Object? isrWithheldCents = null,Object? simultaneous = null,Object? leftBeforeYearEnd = null,Object? payerNotWithholding = null,Object? id = freezed,Object? employerLabel = freezed,}) {
  return _then(_SalaryStatementCreate(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,accumulableCents: null == accumulableCents ? _self.accumulableCents : accumulableCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,exemptCents: null == exemptCents ? _self.exemptCents : exemptCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrWithheldCents: null == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,simultaneous: null == simultaneous ? _self.simultaneous : simultaneous // ignore: cast_nullable_to_non_nullable
as bool,leftBeforeYearEnd: null == leftBeforeYearEnd ? _self.leftBeforeYearEnd : leftBeforeYearEnd // ignore: cast_nullable_to_non_nullable
as bool,payerNotWithholding: null == payerNotWithholding ? _self.payerNotWithholding : payerNotWithholding // ignore: cast_nullable_to_non_nullable
as bool,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,employerLabel: freezed == employerLabel ? _self.employerLabel : employerLabel // ignore: cast_nullable_to_non_nullable
as EmployerLabel?,
  ));
}


}

// dart format on
