// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'salary_statement.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$SalaryStatement {

 Uuid get id;@JsonKey(name: 'taxpayer_id') Uuid get taxpayerId;@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear;@JsonKey(includeIfNull: true, name: 'employer_label') EmployerLabel? get employerLabel;@JsonKey(name: 'accumulable_cents') NonNegativeCents get accumulableCents;@JsonKey(name: 'exempt_cents') NonNegativeCents get exemptCents;@JsonKey(name: 'isr_withheld_cents') NonNegativeCents get isrWithheldCents;/// Worked for two or more employers at the same time.
 bool get simultaneous;@JsonKey(name: 'left_before_year_end') bool get leftBeforeYearEnd;/// The employer was not obliged to withhold (e.g. an international organization).
@JsonKey(name: 'payer_not_withholding') bool get payerNotWithholding; List<Warning> get warnings;@JsonKey(name: 'created_at') DateTime get createdAt;@JsonKey(name: 'updated_at') DateTime get updatedAt;
/// Create a copy of SalaryStatement
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$SalaryStatementCopyWith<SalaryStatement> get copyWith => _$SalaryStatementCopyWithImpl<SalaryStatement>(this as SalaryStatement, _$identity);

  /// Serializes this SalaryStatement to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as SalaryStatement;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is SalaryStatement&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.taxpayerId, _this.taxpayerId) || other.taxpayerId == _this.taxpayerId)&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&(identical(other.employerLabel, _this.employerLabel) || other.employerLabel == _this.employerLabel)&&(identical(other.accumulableCents, _this.accumulableCents) || other.accumulableCents == _this.accumulableCents)&&(identical(other.exemptCents, _this.exemptCents) || other.exemptCents == _this.exemptCents)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.simultaneous, _this.simultaneous) || other.simultaneous == _this.simultaneous)&&(identical(other.leftBeforeYearEnd, _this.leftBeforeYearEnd) || other.leftBeforeYearEnd == _this.leftBeforeYearEnd)&&(identical(other.payerNotWithholding, _this.payerNotWithholding) || other.payerNotWithholding == _this.payerNotWithholding)&&const DeepCollectionEquality().equals(other.warnings, _this.warnings)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.updatedAt, _this.updatedAt) || other.updatedAt == _this.updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as SalaryStatement;
  return Object.hash(runtimeType,_this.id,_this.taxpayerId,_this.fiscalYear,_this.employerLabel,_this.accumulableCents,_this.exemptCents,_this.isrWithheldCents,_this.simultaneous,_this.leftBeforeYearEnd,_this.payerNotWithholding,const DeepCollectionEquality().hash(_this.warnings),_this.createdAt,_this.updatedAt);
}

@override
String toString() {
  final _this = this as SalaryStatement;
  return 'SalaryStatement(id: ${_this.id}, taxpayerId: ${_this.taxpayerId}, fiscalYear: ${_this.fiscalYear}, employerLabel: ${_this.employerLabel}, accumulableCents: ${_this.accumulableCents}, exemptCents: ${_this.exemptCents}, isrWithheldCents: ${_this.isrWithheldCents}, simultaneous: ${_this.simultaneous}, leftBeforeYearEnd: ${_this.leftBeforeYearEnd}, payerNotWithholding: ${_this.payerNotWithholding}, warnings: ${_this.warnings}, createdAt: ${_this.createdAt}, updatedAt: ${_this.updatedAt})';
}


}

/// @nodoc
abstract mixin class $SalaryStatementCopyWith<$Res>  {
  factory $SalaryStatementCopyWith(SalaryStatement value, $Res Function(SalaryStatement) _then) = _$SalaryStatementCopyWithImpl;
@useResult
$Res call({
 Uuid id,@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear,@JsonKey(includeIfNull: true, name: 'employer_label') EmployerLabel? employerLabel,@JsonKey(name: 'accumulable_cents') NonNegativeCents accumulableCents,@JsonKey(name: 'exempt_cents') NonNegativeCents exemptCents,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents isrWithheldCents, bool simultaneous,@JsonKey(name: 'left_before_year_end') bool leftBeforeYearEnd,@JsonKey(name: 'payer_not_withholding') bool payerNotWithholding, List<Warning> warnings,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt
});




}
/// @nodoc
class _$SalaryStatementCopyWithImpl<$Res>
    implements $SalaryStatementCopyWith<$Res> {
  _$SalaryStatementCopyWithImpl(this._self, this._then);

  final SalaryStatement _self;
  final $Res Function(SalaryStatement) _then;

/// Create a copy of SalaryStatement
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? taxpayerId = null,Object? fiscalYear = null,Object? employerLabel = freezed,Object? accumulableCents = null,Object? exemptCents = null,Object? isrWithheldCents = null,Object? simultaneous = null,Object? leftBeforeYearEnd = null,Object? payerNotWithholding = null,Object? warnings = null,Object? createdAt = null,Object? updatedAt = null,}) {
  return _then(SalaryStatement(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,employerLabel: freezed == employerLabel ? _self.employerLabel : employerLabel // ignore: cast_nullable_to_non_nullable
as EmployerLabel?,accumulableCents: null == accumulableCents ? _self.accumulableCents : accumulableCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,exemptCents: null == exemptCents ? _self.exemptCents : exemptCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrWithheldCents: null == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,simultaneous: null == simultaneous ? _self.simultaneous : simultaneous // ignore: cast_nullable_to_non_nullable
as bool,leftBeforeYearEnd: null == leftBeforeYearEnd ? _self.leftBeforeYearEnd : leftBeforeYearEnd // ignore: cast_nullable_to_non_nullable
as bool,payerNotWithholding: null == payerNotWithholding ? _self.payerNotWithholding : payerNotWithholding // ignore: cast_nullable_to_non_nullable
as bool,warnings: null == warnings ? _self.warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}

}


/// Adds pattern-matching-related methods to [SalaryStatement].
extension SalaryStatementPatterns on SalaryStatement {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _SalaryStatement value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _SalaryStatement() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _SalaryStatement value)  $default,){
final _that = this;
switch (_that) {
case _SalaryStatement():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _SalaryStatement value)?  $default,){
final _that = this;
switch (_that) {
case _SalaryStatement() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(includeIfNull: true, name: 'employer_label')  EmployerLabel? employerLabel, @JsonKey(name: 'accumulable_cents')  NonNegativeCents accumulableCents, @JsonKey(name: 'exempt_cents')  NonNegativeCents exemptCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents,  bool simultaneous, @JsonKey(name: 'left_before_year_end')  bool leftBeforeYearEnd, @JsonKey(name: 'payer_not_withholding')  bool payerNotWithholding,  List<Warning> warnings, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _SalaryStatement() when $default != null:
return $default(_that.id,_that.taxpayerId,_that.fiscalYear,_that.employerLabel,_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents,_that.simultaneous,_that.leftBeforeYearEnd,_that.payerNotWithholding,_that.warnings,_that.createdAt,_that.updatedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(includeIfNull: true, name: 'employer_label')  EmployerLabel? employerLabel, @JsonKey(name: 'accumulable_cents')  NonNegativeCents accumulableCents, @JsonKey(name: 'exempt_cents')  NonNegativeCents exemptCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents,  bool simultaneous, @JsonKey(name: 'left_before_year_end')  bool leftBeforeYearEnd, @JsonKey(name: 'payer_not_withholding')  bool payerNotWithholding,  List<Warning> warnings, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)  $default,) {final _that = this;
switch (_that) {
case _SalaryStatement():
return $default(_that.id,_that.taxpayerId,_that.fiscalYear,_that.employerLabel,_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents,_that.simultaneous,_that.leftBeforeYearEnd,_that.payerNotWithholding,_that.warnings,_that.createdAt,_that.updatedAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(includeIfNull: true, name: 'employer_label')  EmployerLabel? employerLabel, @JsonKey(name: 'accumulable_cents')  NonNegativeCents accumulableCents, @JsonKey(name: 'exempt_cents')  NonNegativeCents exemptCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents,  bool simultaneous, @JsonKey(name: 'left_before_year_end')  bool leftBeforeYearEnd, @JsonKey(name: 'payer_not_withholding')  bool payerNotWithholding,  List<Warning> warnings, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,) {final _that = this;
switch (_that) {
case _SalaryStatement() when $default != null:
return $default(_that.id,_that.taxpayerId,_that.fiscalYear,_that.employerLabel,_that.accumulableCents,_that.exemptCents,_that.isrWithheldCents,_that.simultaneous,_that.leftBeforeYearEnd,_that.payerNotWithholding,_that.warnings,_that.createdAt,_that.updatedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _SalaryStatement implements SalaryStatement {
  const _SalaryStatement({required this.id, @JsonKey(name: 'taxpayer_id') required this.taxpayerId, @JsonKey(name: 'fiscal_year') required this.fiscalYear, @JsonKey(includeIfNull: true, name: 'employer_label') required this.employerLabel, @JsonKey(name: 'accumulable_cents') required this.accumulableCents, @JsonKey(name: 'exempt_cents') required this.exemptCents, @JsonKey(name: 'isr_withheld_cents') required this.isrWithheldCents, required this.simultaneous, @JsonKey(name: 'left_before_year_end') required this.leftBeforeYearEnd, @JsonKey(name: 'payer_not_withholding') required this.payerNotWithholding, required  List<Warning> warnings, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(name: 'updated_at') required this.updatedAt}): _warnings = warnings;
  factory _SalaryStatement.fromJson(Map<String, dynamic> json) => _$SalaryStatementFromJson(json);

@override final  Uuid id;
@override@JsonKey(name: 'taxpayer_id') final  Uuid taxpayerId;
@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
@override@JsonKey(includeIfNull: true, name: 'employer_label') final  EmployerLabel? employerLabel;
@override@JsonKey(name: 'accumulable_cents') final  NonNegativeCents accumulableCents;
@override@JsonKey(name: 'exempt_cents') final  NonNegativeCents exemptCents;
@override@JsonKey(name: 'isr_withheld_cents') final  NonNegativeCents isrWithheldCents;
/// Worked for two or more employers at the same time.
@override final  bool simultaneous;
@override@JsonKey(name: 'left_before_year_end') final  bool leftBeforeYearEnd;
/// The employer was not obliged to withhold (e.g. an international organization).
@override@JsonKey(name: 'payer_not_withholding') final  bool payerNotWithholding;
 final  List<Warning> _warnings;
@override List<Warning> get warnings {
  if (_warnings is EqualUnmodifiableListView) return _warnings;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_warnings);
}

@override@JsonKey(name: 'created_at') final  DateTime createdAt;
@override@JsonKey(name: 'updated_at') final  DateTime updatedAt;

/// Create a copy of SalaryStatement
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$SalaryStatementCopyWith<_SalaryStatement> get copyWith => __$SalaryStatementCopyWithImpl<_SalaryStatement>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$SalaryStatementToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _SalaryStatement&&(identical(other.id, id) || other.id == id)&&(identical(other.taxpayerId, taxpayerId) || other.taxpayerId == taxpayerId)&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&(identical(other.employerLabel, employerLabel) || other.employerLabel == employerLabel)&&(identical(other.accumulableCents, accumulableCents) || other.accumulableCents == accumulableCents)&&(identical(other.exemptCents, exemptCents) || other.exemptCents == exemptCents)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.simultaneous, simultaneous) || other.simultaneous == simultaneous)&&(identical(other.leftBeforeYearEnd, leftBeforeYearEnd) || other.leftBeforeYearEnd == leftBeforeYearEnd)&&(identical(other.payerNotWithholding, payerNotWithholding) || other.payerNotWithholding == payerNotWithholding)&&const DeepCollectionEquality().equals(other.warnings, _warnings)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,taxpayerId,fiscalYear,employerLabel,accumulableCents,exemptCents,isrWithheldCents,simultaneous,leftBeforeYearEnd,payerNotWithholding,const DeepCollectionEquality().hash(_warnings),createdAt,updatedAt);
}

@override
String toString() {
    return 'SalaryStatement(id: $id, taxpayerId: $taxpayerId, fiscalYear: $fiscalYear, employerLabel: $employerLabel, accumulableCents: $accumulableCents, exemptCents: $exemptCents, isrWithheldCents: $isrWithheldCents, simultaneous: $simultaneous, leftBeforeYearEnd: $leftBeforeYearEnd, payerNotWithholding: $payerNotWithholding, warnings: $warnings, createdAt: $createdAt, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class _$SalaryStatementCopyWith<$Res> implements $SalaryStatementCopyWith<$Res> {
  factory _$SalaryStatementCopyWith(_SalaryStatement value, $Res Function(_SalaryStatement) _then) = __$SalaryStatementCopyWithImpl;
@override @useResult
$Res call({
 Uuid id,@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear,@JsonKey(includeIfNull: true, name: 'employer_label') EmployerLabel? employerLabel,@JsonKey(name: 'accumulable_cents') NonNegativeCents accumulableCents,@JsonKey(name: 'exempt_cents') NonNegativeCents exemptCents,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents isrWithheldCents, bool simultaneous,@JsonKey(name: 'left_before_year_end') bool leftBeforeYearEnd,@JsonKey(name: 'payer_not_withholding') bool payerNotWithholding, List<Warning> warnings,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt
});




}
/// @nodoc
class __$SalaryStatementCopyWithImpl<$Res>
    implements _$SalaryStatementCopyWith<$Res> {
  __$SalaryStatementCopyWithImpl(this._self, this._then);

  final _SalaryStatement _self;
  final $Res Function(_SalaryStatement) _then;

/// Create a copy of SalaryStatement
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? taxpayerId = null,Object? fiscalYear = null,Object? employerLabel = freezed,Object? accumulableCents = null,Object? exemptCents = null,Object? isrWithheldCents = null,Object? simultaneous = null,Object? leftBeforeYearEnd = null,Object? payerNotWithholding = null,Object? warnings = null,Object? createdAt = null,Object? updatedAt = null,}) {
  return _then(_SalaryStatement(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,employerLabel: freezed == employerLabel ? _self.employerLabel : employerLabel // ignore: cast_nullable_to_non_nullable
as EmployerLabel?,accumulableCents: null == accumulableCents ? _self.accumulableCents : accumulableCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,exemptCents: null == exemptCents ? _self.exemptCents : exemptCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrWithheldCents: null == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,simultaneous: null == simultaneous ? _self.simultaneous : simultaneous // ignore: cast_nullable_to_non_nullable
as bool,leftBeforeYearEnd: null == leftBeforeYearEnd ? _self.leftBeforeYearEnd : leftBeforeYearEnd // ignore: cast_nullable_to_non_nullable
as bool,payerNotWithholding: null == payerNotWithholding ? _self.payerNotWithholding : payerNotWithholding // ignore: cast_nullable_to_non_nullable
as bool,warnings: null == warnings ? _self._warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}


}

// dart format on
