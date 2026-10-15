// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'personal_deduction.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$PersonalDeduction {

 Uuid get id;@JsonKey(name: 'taxpayer_id') Uuid get taxpayerId;@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear; DeductionCategory get category;@JsonKey(name: 'amount_cents') PositiveCents get amountCents;@JsonKey(name: 'paid_on') String get paidOn;@JsonKey(name: 'payment_method') DeductionPaymentMethod get paymentMethod; DeductionBeneficiary get beneficiary;@JsonKey(includeIfNull: true, name: 'donee_type') DoneeType? get doneeType;@JsonKey(includeIfNull: true, name: 'student_label') StudentLabel? get studentLabel;@JsonKey(includeIfNull: true, name: 'tuition_level') TuitionLevel? get tuitionLevel;@JsonKey(includeIfNull: true, name: 'scholarship_cents') NonNegativeCents? get scholarshipCents;@JsonKey(includeIfNull: true, name: 'cfdi_uuid') CfdiUuid? get cfdiUuid;@JsonKey(includeIfNull: true) DeductionExclusion? get excluded; List<Warning> get warnings;@JsonKey(name: 'created_at') DateTime get createdAt;@JsonKey(name: 'updated_at') DateTime get updatedAt;
/// Create a copy of PersonalDeduction
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$PersonalDeductionCopyWith<PersonalDeduction> get copyWith => _$PersonalDeductionCopyWithImpl<PersonalDeduction>(this as PersonalDeduction, _$identity);

  /// Serializes this PersonalDeduction to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as PersonalDeduction;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is PersonalDeduction&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.taxpayerId, _this.taxpayerId) || other.taxpayerId == _this.taxpayerId)&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&(identical(other.category, _this.category) || other.category == _this.category)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.paidOn, _this.paidOn) || other.paidOn == _this.paidOn)&&(identical(other.paymentMethod, _this.paymentMethod) || other.paymentMethod == _this.paymentMethod)&&(identical(other.beneficiary, _this.beneficiary) || other.beneficiary == _this.beneficiary)&&(identical(other.doneeType, _this.doneeType) || other.doneeType == _this.doneeType)&&(identical(other.studentLabel, _this.studentLabel) || other.studentLabel == _this.studentLabel)&&(identical(other.tuitionLevel, _this.tuitionLevel) || other.tuitionLevel == _this.tuitionLevel)&&(identical(other.scholarshipCents, _this.scholarshipCents) || other.scholarshipCents == _this.scholarshipCents)&&(identical(other.cfdiUuid, _this.cfdiUuid) || other.cfdiUuid == _this.cfdiUuid)&&(identical(other.excluded, _this.excluded) || other.excluded == _this.excluded)&&const DeepCollectionEquality().equals(other.warnings, _this.warnings)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.updatedAt, _this.updatedAt) || other.updatedAt == _this.updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as PersonalDeduction;
  return Object.hash(runtimeType,_this.id,_this.taxpayerId,_this.fiscalYear,_this.category,_this.amountCents,_this.paidOn,_this.paymentMethod,_this.beneficiary,_this.doneeType,_this.studentLabel,_this.tuitionLevel,_this.scholarshipCents,_this.cfdiUuid,_this.excluded,const DeepCollectionEquality().hash(_this.warnings),_this.createdAt,_this.updatedAt);
}

@override
String toString() {
  final _this = this as PersonalDeduction;
  return 'PersonalDeduction(id: ${_this.id}, taxpayerId: ${_this.taxpayerId}, fiscalYear: ${_this.fiscalYear}, category: ${_this.category}, amountCents: ${_this.amountCents}, paidOn: ${_this.paidOn}, paymentMethod: ${_this.paymentMethod}, beneficiary: ${_this.beneficiary}, doneeType: ${_this.doneeType}, studentLabel: ${_this.studentLabel}, tuitionLevel: ${_this.tuitionLevel}, scholarshipCents: ${_this.scholarshipCents}, cfdiUuid: ${_this.cfdiUuid}, excluded: ${_this.excluded}, warnings: ${_this.warnings}, createdAt: ${_this.createdAt}, updatedAt: ${_this.updatedAt})';
}


}

/// @nodoc
abstract mixin class $PersonalDeductionCopyWith<$Res>  {
  factory $PersonalDeductionCopyWith(PersonalDeduction value, $Res Function(PersonalDeduction) _then) = _$PersonalDeductionCopyWithImpl;
@useResult
$Res call({
 Uuid id,@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, DeductionCategory category,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'paid_on') String paidOn,@JsonKey(name: 'payment_method') DeductionPaymentMethod paymentMethod, DeductionBeneficiary beneficiary,@JsonKey(includeIfNull: true, name: 'donee_type') DoneeType? doneeType,@JsonKey(includeIfNull: true, name: 'student_label') StudentLabel? studentLabel,@JsonKey(includeIfNull: true, name: 'tuition_level') TuitionLevel? tuitionLevel,@JsonKey(includeIfNull: true, name: 'scholarship_cents') NonNegativeCents? scholarshipCents,@JsonKey(includeIfNull: true, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,@JsonKey(includeIfNull: true) DeductionExclusion? excluded, List<Warning> warnings,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt
});




}
/// @nodoc
class _$PersonalDeductionCopyWithImpl<$Res>
    implements $PersonalDeductionCopyWith<$Res> {
  _$PersonalDeductionCopyWithImpl(this._self, this._then);

  final PersonalDeduction _self;
  final $Res Function(PersonalDeduction) _then;

/// Create a copy of PersonalDeduction
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? taxpayerId = null,Object? fiscalYear = null,Object? category = null,Object? amountCents = null,Object? paidOn = null,Object? paymentMethod = null,Object? beneficiary = null,Object? doneeType = freezed,Object? studentLabel = freezed,Object? tuitionLevel = freezed,Object? scholarshipCents = freezed,Object? cfdiUuid = freezed,Object? excluded = freezed,Object? warnings = null,Object? createdAt = null,Object? updatedAt = null,}) {
  return _then(PersonalDeduction(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as DeductionCategory,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,paidOn: null == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as String,paymentMethod: null == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as DeductionPaymentMethod,beneficiary: null == beneficiary ? _self.beneficiary : beneficiary // ignore: cast_nullable_to_non_nullable
as DeductionBeneficiary,doneeType: freezed == doneeType ? _self.doneeType : doneeType // ignore: cast_nullable_to_non_nullable
as DoneeType?,studentLabel: freezed == studentLabel ? _self.studentLabel : studentLabel // ignore: cast_nullable_to_non_nullable
as StudentLabel?,tuitionLevel: freezed == tuitionLevel ? _self.tuitionLevel : tuitionLevel // ignore: cast_nullable_to_non_nullable
as TuitionLevel?,scholarshipCents: freezed == scholarshipCents ? _self.scholarshipCents : scholarshipCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,excluded: freezed == excluded ? _self.excluded : excluded // ignore: cast_nullable_to_non_nullable
as DeductionExclusion?,warnings: null == warnings ? _self.warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}

}


/// Adds pattern-matching-related methods to [PersonalDeduction].
extension PersonalDeductionPatterns on PersonalDeduction {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _PersonalDeduction value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _PersonalDeduction() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _PersonalDeduction value)  $default,){
final _that = this;
switch (_that) {
case _PersonalDeduction():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _PersonalDeduction value)?  $default,){
final _that = this;
switch (_that) {
case _PersonalDeduction() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  DeductionCategory category, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'paid_on')  String paidOn, @JsonKey(name: 'payment_method')  DeductionPaymentMethod paymentMethod,  DeductionBeneficiary beneficiary, @JsonKey(includeIfNull: true, name: 'donee_type')  DoneeType? doneeType, @JsonKey(includeIfNull: true, name: 'student_label')  StudentLabel? studentLabel, @JsonKey(includeIfNull: true, name: 'tuition_level')  TuitionLevel? tuitionLevel, @JsonKey(includeIfNull: true, name: 'scholarship_cents')  NonNegativeCents? scholarshipCents, @JsonKey(includeIfNull: true, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: true)  DeductionExclusion? excluded,  List<Warning> warnings, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _PersonalDeduction() when $default != null:
return $default(_that.id,_that.taxpayerId,_that.fiscalYear,_that.category,_that.amountCents,_that.paidOn,_that.paymentMethod,_that.beneficiary,_that.doneeType,_that.studentLabel,_that.tuitionLevel,_that.scholarshipCents,_that.cfdiUuid,_that.excluded,_that.warnings,_that.createdAt,_that.updatedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  DeductionCategory category, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'paid_on')  String paidOn, @JsonKey(name: 'payment_method')  DeductionPaymentMethod paymentMethod,  DeductionBeneficiary beneficiary, @JsonKey(includeIfNull: true, name: 'donee_type')  DoneeType? doneeType, @JsonKey(includeIfNull: true, name: 'student_label')  StudentLabel? studentLabel, @JsonKey(includeIfNull: true, name: 'tuition_level')  TuitionLevel? tuitionLevel, @JsonKey(includeIfNull: true, name: 'scholarship_cents')  NonNegativeCents? scholarshipCents, @JsonKey(includeIfNull: true, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: true)  DeductionExclusion? excluded,  List<Warning> warnings, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)  $default,) {final _that = this;
switch (_that) {
case _PersonalDeduction():
return $default(_that.id,_that.taxpayerId,_that.fiscalYear,_that.category,_that.amountCents,_that.paidOn,_that.paymentMethod,_that.beneficiary,_that.doneeType,_that.studentLabel,_that.tuitionLevel,_that.scholarshipCents,_that.cfdiUuid,_that.excluded,_that.warnings,_that.createdAt,_that.updatedAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  DeductionCategory category, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'paid_on')  String paidOn, @JsonKey(name: 'payment_method')  DeductionPaymentMethod paymentMethod,  DeductionBeneficiary beneficiary, @JsonKey(includeIfNull: true, name: 'donee_type')  DoneeType? doneeType, @JsonKey(includeIfNull: true, name: 'student_label')  StudentLabel? studentLabel, @JsonKey(includeIfNull: true, name: 'tuition_level')  TuitionLevel? tuitionLevel, @JsonKey(includeIfNull: true, name: 'scholarship_cents')  NonNegativeCents? scholarshipCents, @JsonKey(includeIfNull: true, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: true)  DeductionExclusion? excluded,  List<Warning> warnings, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,) {final _that = this;
switch (_that) {
case _PersonalDeduction() when $default != null:
return $default(_that.id,_that.taxpayerId,_that.fiscalYear,_that.category,_that.amountCents,_that.paidOn,_that.paymentMethod,_that.beneficiary,_that.doneeType,_that.studentLabel,_that.tuitionLevel,_that.scholarshipCents,_that.cfdiUuid,_that.excluded,_that.warnings,_that.createdAt,_that.updatedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _PersonalDeduction implements PersonalDeduction {
  const _PersonalDeduction({required this.id, @JsonKey(name: 'taxpayer_id') required this.taxpayerId, @JsonKey(name: 'fiscal_year') required this.fiscalYear, required this.category, @JsonKey(name: 'amount_cents') required this.amountCents, @JsonKey(name: 'paid_on') required this.paidOn, @JsonKey(name: 'payment_method') required this.paymentMethod, required this.beneficiary, @JsonKey(includeIfNull: true, name: 'donee_type') required this.doneeType, @JsonKey(includeIfNull: true, name: 'student_label') required this.studentLabel, @JsonKey(includeIfNull: true, name: 'tuition_level') required this.tuitionLevel, @JsonKey(includeIfNull: true, name: 'scholarship_cents') required this.scholarshipCents, @JsonKey(includeIfNull: true, name: 'cfdi_uuid') required this.cfdiUuid, @JsonKey(includeIfNull: true) required this.excluded, required  List<Warning> warnings, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(name: 'updated_at') required this.updatedAt}): _warnings = warnings;
  factory _PersonalDeduction.fromJson(Map<String, dynamic> json) => _$PersonalDeductionFromJson(json);

@override final  Uuid id;
@override@JsonKey(name: 'taxpayer_id') final  Uuid taxpayerId;
@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
@override final  DeductionCategory category;
@override@JsonKey(name: 'amount_cents') final  PositiveCents amountCents;
@override@JsonKey(name: 'paid_on') final  String paidOn;
@override@JsonKey(name: 'payment_method') final  DeductionPaymentMethod paymentMethod;
@override final  DeductionBeneficiary beneficiary;
@override@JsonKey(includeIfNull: true, name: 'donee_type') final  DoneeType? doneeType;
@override@JsonKey(includeIfNull: true, name: 'student_label') final  StudentLabel? studentLabel;
@override@JsonKey(includeIfNull: true, name: 'tuition_level') final  TuitionLevel? tuitionLevel;
@override@JsonKey(includeIfNull: true, name: 'scholarship_cents') final  NonNegativeCents? scholarshipCents;
@override@JsonKey(includeIfNull: true, name: 'cfdi_uuid') final  CfdiUuid? cfdiUuid;
@override@JsonKey(includeIfNull: true) final  DeductionExclusion? excluded;
 final  List<Warning> _warnings;
@override List<Warning> get warnings {
  if (_warnings is EqualUnmodifiableListView) return _warnings;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_warnings);
}

@override@JsonKey(name: 'created_at') final  DateTime createdAt;
@override@JsonKey(name: 'updated_at') final  DateTime updatedAt;

/// Create a copy of PersonalDeduction
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$PersonalDeductionCopyWith<_PersonalDeduction> get copyWith => __$PersonalDeductionCopyWithImpl<_PersonalDeduction>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$PersonalDeductionToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _PersonalDeduction&&(identical(other.id, id) || other.id == id)&&(identical(other.taxpayerId, taxpayerId) || other.taxpayerId == taxpayerId)&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&(identical(other.category, category) || other.category == category)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.paidOn, paidOn) || other.paidOn == paidOn)&&(identical(other.paymentMethod, paymentMethod) || other.paymentMethod == paymentMethod)&&(identical(other.beneficiary, beneficiary) || other.beneficiary == beneficiary)&&(identical(other.doneeType, doneeType) || other.doneeType == doneeType)&&(identical(other.studentLabel, studentLabel) || other.studentLabel == studentLabel)&&(identical(other.tuitionLevel, tuitionLevel) || other.tuitionLevel == tuitionLevel)&&(identical(other.scholarshipCents, scholarshipCents) || other.scholarshipCents == scholarshipCents)&&(identical(other.cfdiUuid, cfdiUuid) || other.cfdiUuid == cfdiUuid)&&(identical(other.excluded, excluded) || other.excluded == excluded)&&const DeepCollectionEquality().equals(other.warnings, _warnings)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,taxpayerId,fiscalYear,category,amountCents,paidOn,paymentMethod,beneficiary,doneeType,studentLabel,tuitionLevel,scholarshipCents,cfdiUuid,excluded,const DeepCollectionEquality().hash(_warnings),createdAt,updatedAt);
}

@override
String toString() {
    return 'PersonalDeduction(id: $id, taxpayerId: $taxpayerId, fiscalYear: $fiscalYear, category: $category, amountCents: $amountCents, paidOn: $paidOn, paymentMethod: $paymentMethod, beneficiary: $beneficiary, doneeType: $doneeType, studentLabel: $studentLabel, tuitionLevel: $tuitionLevel, scholarshipCents: $scholarshipCents, cfdiUuid: $cfdiUuid, excluded: $excluded, warnings: $warnings, createdAt: $createdAt, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class _$PersonalDeductionCopyWith<$Res> implements $PersonalDeductionCopyWith<$Res> {
  factory _$PersonalDeductionCopyWith(_PersonalDeduction value, $Res Function(_PersonalDeduction) _then) = __$PersonalDeductionCopyWithImpl;
@override @useResult
$Res call({
 Uuid id,@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, DeductionCategory category,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'paid_on') String paidOn,@JsonKey(name: 'payment_method') DeductionPaymentMethod paymentMethod, DeductionBeneficiary beneficiary,@JsonKey(includeIfNull: true, name: 'donee_type') DoneeType? doneeType,@JsonKey(includeIfNull: true, name: 'student_label') StudentLabel? studentLabel,@JsonKey(includeIfNull: true, name: 'tuition_level') TuitionLevel? tuitionLevel,@JsonKey(includeIfNull: true, name: 'scholarship_cents') NonNegativeCents? scholarshipCents,@JsonKey(includeIfNull: true, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,@JsonKey(includeIfNull: true) DeductionExclusion? excluded, List<Warning> warnings,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt
});




}
/// @nodoc
class __$PersonalDeductionCopyWithImpl<$Res>
    implements _$PersonalDeductionCopyWith<$Res> {
  __$PersonalDeductionCopyWithImpl(this._self, this._then);

  final _PersonalDeduction _self;
  final $Res Function(_PersonalDeduction) _then;

/// Create a copy of PersonalDeduction
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? taxpayerId = null,Object? fiscalYear = null,Object? category = null,Object? amountCents = null,Object? paidOn = null,Object? paymentMethod = null,Object? beneficiary = null,Object? doneeType = freezed,Object? studentLabel = freezed,Object? tuitionLevel = freezed,Object? scholarshipCents = freezed,Object? cfdiUuid = freezed,Object? excluded = freezed,Object? warnings = null,Object? createdAt = null,Object? updatedAt = null,}) {
  return _then(_PersonalDeduction(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as DeductionCategory,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,paidOn: null == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as String,paymentMethod: null == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as DeductionPaymentMethod,beneficiary: null == beneficiary ? _self.beneficiary : beneficiary // ignore: cast_nullable_to_non_nullable
as DeductionBeneficiary,doneeType: freezed == doneeType ? _self.doneeType : doneeType // ignore: cast_nullable_to_non_nullable
as DoneeType?,studentLabel: freezed == studentLabel ? _self.studentLabel : studentLabel // ignore: cast_nullable_to_non_nullable
as StudentLabel?,tuitionLevel: freezed == tuitionLevel ? _self.tuitionLevel : tuitionLevel // ignore: cast_nullable_to_non_nullable
as TuitionLevel?,scholarshipCents: freezed == scholarshipCents ? _self.scholarshipCents : scholarshipCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,excluded: freezed == excluded ? _self.excluded : excluded // ignore: cast_nullable_to_non_nullable
as DeductionExclusion?,warnings: null == warnings ? _self._warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}


}

// dart format on
