// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'personal_deduction_create.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$PersonalDeductionCreate {

 DeductionCategory get category;@JsonKey(name: 'amount_cents') PositiveCents get amountCents;@JsonKey(name: 'paid_on') String get paidOn;@JsonKey(name: 'payment_method') DeductionPaymentMethod get paymentMethod; DeductionBeneficiary get beneficiary;@JsonKey(includeIfNull: false) Uuid? get id;@JsonKey(includeIfNull: false, name: 'donee_type') DoneeType? get doneeType;@JsonKey(includeIfNull: false, name: 'student_label') StudentLabel? get studentLabel;@JsonKey(includeIfNull: false, name: 'tuition_level') TuitionLevel? get tuitionLevel;@JsonKey(includeIfNull: false, name: 'scholarship_cents') NonNegativeCents? get scholarshipCents;@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? get cfdiUuid;
/// Create a copy of PersonalDeductionCreate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$PersonalDeductionCreateCopyWith<PersonalDeductionCreate> get copyWith => _$PersonalDeductionCreateCopyWithImpl<PersonalDeductionCreate>(this as PersonalDeductionCreate, _$identity);

  /// Serializes this PersonalDeductionCreate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as PersonalDeductionCreate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is PersonalDeductionCreate&&(identical(other.category, _this.category) || other.category == _this.category)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.paidOn, _this.paidOn) || other.paidOn == _this.paidOn)&&(identical(other.paymentMethod, _this.paymentMethod) || other.paymentMethod == _this.paymentMethod)&&(identical(other.beneficiary, _this.beneficiary) || other.beneficiary == _this.beneficiary)&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.doneeType, _this.doneeType) || other.doneeType == _this.doneeType)&&(identical(other.studentLabel, _this.studentLabel) || other.studentLabel == _this.studentLabel)&&(identical(other.tuitionLevel, _this.tuitionLevel) || other.tuitionLevel == _this.tuitionLevel)&&(identical(other.scholarshipCents, _this.scholarshipCents) || other.scholarshipCents == _this.scholarshipCents)&&(identical(other.cfdiUuid, _this.cfdiUuid) || other.cfdiUuid == _this.cfdiUuid));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as PersonalDeductionCreate;
  return Object.hash(runtimeType,_this.category,_this.amountCents,_this.paidOn,_this.paymentMethod,_this.beneficiary,_this.id,_this.doneeType,_this.studentLabel,_this.tuitionLevel,_this.scholarshipCents,_this.cfdiUuid);
}

@override
String toString() {
  final _this = this as PersonalDeductionCreate;
  return 'PersonalDeductionCreate(category: ${_this.category}, amountCents: ${_this.amountCents}, paidOn: ${_this.paidOn}, paymentMethod: ${_this.paymentMethod}, beneficiary: ${_this.beneficiary}, id: ${_this.id}, doneeType: ${_this.doneeType}, studentLabel: ${_this.studentLabel}, tuitionLevel: ${_this.tuitionLevel}, scholarshipCents: ${_this.scholarshipCents}, cfdiUuid: ${_this.cfdiUuid})';
}


}

/// @nodoc
abstract mixin class $PersonalDeductionCreateCopyWith<$Res>  {
  factory $PersonalDeductionCreateCopyWith(PersonalDeductionCreate value, $Res Function(PersonalDeductionCreate) _then) = _$PersonalDeductionCreateCopyWithImpl;
@useResult
$Res call({
 DeductionCategory category,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'paid_on') String paidOn,@JsonKey(name: 'payment_method') DeductionPaymentMethod paymentMethod, DeductionBeneficiary beneficiary,@JsonKey(includeIfNull: false) Uuid? id,@JsonKey(includeIfNull: false, name: 'donee_type') DoneeType? doneeType,@JsonKey(includeIfNull: false, name: 'student_label') StudentLabel? studentLabel,@JsonKey(includeIfNull: false, name: 'tuition_level') TuitionLevel? tuitionLevel,@JsonKey(includeIfNull: false, name: 'scholarship_cents') NonNegativeCents? scholarshipCents,@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid
});




}
/// @nodoc
class _$PersonalDeductionCreateCopyWithImpl<$Res>
    implements $PersonalDeductionCreateCopyWith<$Res> {
  _$PersonalDeductionCreateCopyWithImpl(this._self, this._then);

  final PersonalDeductionCreate _self;
  final $Res Function(PersonalDeductionCreate) _then;

/// Create a copy of PersonalDeductionCreate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? category = null,Object? amountCents = null,Object? paidOn = null,Object? paymentMethod = null,Object? beneficiary = null,Object? id = freezed,Object? doneeType = freezed,Object? studentLabel = freezed,Object? tuitionLevel = freezed,Object? scholarshipCents = freezed,Object? cfdiUuid = freezed,}) {
  return _then(PersonalDeductionCreate(
category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as DeductionCategory,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,paidOn: null == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as String,paymentMethod: null == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as DeductionPaymentMethod,beneficiary: null == beneficiary ? _self.beneficiary : beneficiary // ignore: cast_nullable_to_non_nullable
as DeductionBeneficiary,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,doneeType: freezed == doneeType ? _self.doneeType : doneeType // ignore: cast_nullable_to_non_nullable
as DoneeType?,studentLabel: freezed == studentLabel ? _self.studentLabel : studentLabel // ignore: cast_nullable_to_non_nullable
as StudentLabel?,tuitionLevel: freezed == tuitionLevel ? _self.tuitionLevel : tuitionLevel // ignore: cast_nullable_to_non_nullable
as TuitionLevel?,scholarshipCents: freezed == scholarshipCents ? _self.scholarshipCents : scholarshipCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,
  ));
}

}


/// Adds pattern-matching-related methods to [PersonalDeductionCreate].
extension PersonalDeductionCreatePatterns on PersonalDeductionCreate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _PersonalDeductionCreate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _PersonalDeductionCreate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _PersonalDeductionCreate value)  $default,){
final _that = this;
switch (_that) {
case _PersonalDeductionCreate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _PersonalDeductionCreate value)?  $default,){
final _that = this;
switch (_that) {
case _PersonalDeductionCreate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( DeductionCategory category, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'paid_on')  String paidOn, @JsonKey(name: 'payment_method')  DeductionPaymentMethod paymentMethod,  DeductionBeneficiary beneficiary, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'donee_type')  DoneeType? doneeType, @JsonKey(includeIfNull: false, name: 'student_label')  StudentLabel? studentLabel, @JsonKey(includeIfNull: false, name: 'tuition_level')  TuitionLevel? tuitionLevel, @JsonKey(includeIfNull: false, name: 'scholarship_cents')  NonNegativeCents? scholarshipCents, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _PersonalDeductionCreate() when $default != null:
return $default(_that.category,_that.amountCents,_that.paidOn,_that.paymentMethod,_that.beneficiary,_that.id,_that.doneeType,_that.studentLabel,_that.tuitionLevel,_that.scholarshipCents,_that.cfdiUuid);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( DeductionCategory category, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'paid_on')  String paidOn, @JsonKey(name: 'payment_method')  DeductionPaymentMethod paymentMethod,  DeductionBeneficiary beneficiary, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'donee_type')  DoneeType? doneeType, @JsonKey(includeIfNull: false, name: 'student_label')  StudentLabel? studentLabel, @JsonKey(includeIfNull: false, name: 'tuition_level')  TuitionLevel? tuitionLevel, @JsonKey(includeIfNull: false, name: 'scholarship_cents')  NonNegativeCents? scholarshipCents, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid)  $default,) {final _that = this;
switch (_that) {
case _PersonalDeductionCreate():
return $default(_that.category,_that.amountCents,_that.paidOn,_that.paymentMethod,_that.beneficiary,_that.id,_that.doneeType,_that.studentLabel,_that.tuitionLevel,_that.scholarshipCents,_that.cfdiUuid);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( DeductionCategory category, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'paid_on')  String paidOn, @JsonKey(name: 'payment_method')  DeductionPaymentMethod paymentMethod,  DeductionBeneficiary beneficiary, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'donee_type')  DoneeType? doneeType, @JsonKey(includeIfNull: false, name: 'student_label')  StudentLabel? studentLabel, @JsonKey(includeIfNull: false, name: 'tuition_level')  TuitionLevel? tuitionLevel, @JsonKey(includeIfNull: false, name: 'scholarship_cents')  NonNegativeCents? scholarshipCents, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid)?  $default,) {final _that = this;
switch (_that) {
case _PersonalDeductionCreate() when $default != null:
return $default(_that.category,_that.amountCents,_that.paidOn,_that.paymentMethod,_that.beneficiary,_that.id,_that.doneeType,_that.studentLabel,_that.tuitionLevel,_that.scholarshipCents,_that.cfdiUuid);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _PersonalDeductionCreate implements PersonalDeductionCreate {
  const _PersonalDeductionCreate({required this.category, @JsonKey(name: 'amount_cents') required this.amountCents, @JsonKey(name: 'paid_on') required this.paidOn, @JsonKey(name: 'payment_method') required this.paymentMethod, required this.beneficiary, @JsonKey(includeIfNull: false) this.id, @JsonKey(includeIfNull: false, name: 'donee_type') this.doneeType, @JsonKey(includeIfNull: false, name: 'student_label') this.studentLabel, @JsonKey(includeIfNull: false, name: 'tuition_level') this.tuitionLevel, @JsonKey(includeIfNull: false, name: 'scholarship_cents') this.scholarshipCents, @JsonKey(includeIfNull: false, name: 'cfdi_uuid') this.cfdiUuid});
  factory _PersonalDeductionCreate.fromJson(Map<String, dynamic> json) => _$PersonalDeductionCreateFromJson(json);

@override final  DeductionCategory category;
@override@JsonKey(name: 'amount_cents') final  PositiveCents amountCents;
@override@JsonKey(name: 'paid_on') final  String paidOn;
@override@JsonKey(name: 'payment_method') final  DeductionPaymentMethod paymentMethod;
@override final  DeductionBeneficiary beneficiary;
@override@JsonKey(includeIfNull: false) final  Uuid? id;
@override@JsonKey(includeIfNull: false, name: 'donee_type') final  DoneeType? doneeType;
@override@JsonKey(includeIfNull: false, name: 'student_label') final  StudentLabel? studentLabel;
@override@JsonKey(includeIfNull: false, name: 'tuition_level') final  TuitionLevel? tuitionLevel;
@override@JsonKey(includeIfNull: false, name: 'scholarship_cents') final  NonNegativeCents? scholarshipCents;
@override@JsonKey(includeIfNull: false, name: 'cfdi_uuid') final  CfdiUuid? cfdiUuid;

/// Create a copy of PersonalDeductionCreate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$PersonalDeductionCreateCopyWith<_PersonalDeductionCreate> get copyWith => __$PersonalDeductionCreateCopyWithImpl<_PersonalDeductionCreate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$PersonalDeductionCreateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _PersonalDeductionCreate&&(identical(other.category, category) || other.category == category)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.paidOn, paidOn) || other.paidOn == paidOn)&&(identical(other.paymentMethod, paymentMethod) || other.paymentMethod == paymentMethod)&&(identical(other.beneficiary, beneficiary) || other.beneficiary == beneficiary)&&(identical(other.id, id) || other.id == id)&&(identical(other.doneeType, doneeType) || other.doneeType == doneeType)&&(identical(other.studentLabel, studentLabel) || other.studentLabel == studentLabel)&&(identical(other.tuitionLevel, tuitionLevel) || other.tuitionLevel == tuitionLevel)&&(identical(other.scholarshipCents, scholarshipCents) || other.scholarshipCents == scholarshipCents)&&(identical(other.cfdiUuid, cfdiUuid) || other.cfdiUuid == cfdiUuid));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,category,amountCents,paidOn,paymentMethod,beneficiary,id,doneeType,studentLabel,tuitionLevel,scholarshipCents,cfdiUuid);
}

@override
String toString() {
    return 'PersonalDeductionCreate(category: $category, amountCents: $amountCents, paidOn: $paidOn, paymentMethod: $paymentMethod, beneficiary: $beneficiary, id: $id, doneeType: $doneeType, studentLabel: $studentLabel, tuitionLevel: $tuitionLevel, scholarshipCents: $scholarshipCents, cfdiUuid: $cfdiUuid)';
}


}

/// @nodoc
abstract mixin class _$PersonalDeductionCreateCopyWith<$Res> implements $PersonalDeductionCreateCopyWith<$Res> {
  factory _$PersonalDeductionCreateCopyWith(_PersonalDeductionCreate value, $Res Function(_PersonalDeductionCreate) _then) = __$PersonalDeductionCreateCopyWithImpl;
@override @useResult
$Res call({
 DeductionCategory category,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'paid_on') String paidOn,@JsonKey(name: 'payment_method') DeductionPaymentMethod paymentMethod, DeductionBeneficiary beneficiary,@JsonKey(includeIfNull: false) Uuid? id,@JsonKey(includeIfNull: false, name: 'donee_type') DoneeType? doneeType,@JsonKey(includeIfNull: false, name: 'student_label') StudentLabel? studentLabel,@JsonKey(includeIfNull: false, name: 'tuition_level') TuitionLevel? tuitionLevel,@JsonKey(includeIfNull: false, name: 'scholarship_cents') NonNegativeCents? scholarshipCents,@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid
});




}
/// @nodoc
class __$PersonalDeductionCreateCopyWithImpl<$Res>
    implements _$PersonalDeductionCreateCopyWith<$Res> {
  __$PersonalDeductionCreateCopyWithImpl(this._self, this._then);

  final _PersonalDeductionCreate _self;
  final $Res Function(_PersonalDeductionCreate) _then;

/// Create a copy of PersonalDeductionCreate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? category = null,Object? amountCents = null,Object? paidOn = null,Object? paymentMethod = null,Object? beneficiary = null,Object? id = freezed,Object? doneeType = freezed,Object? studentLabel = freezed,Object? tuitionLevel = freezed,Object? scholarshipCents = freezed,Object? cfdiUuid = freezed,}) {
  return _then(_PersonalDeductionCreate(
category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as DeductionCategory,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,paidOn: null == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as String,paymentMethod: null == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as DeductionPaymentMethod,beneficiary: null == beneficiary ? _self.beneficiary : beneficiary // ignore: cast_nullable_to_non_nullable
as DeductionBeneficiary,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,doneeType: freezed == doneeType ? _self.doneeType : doneeType // ignore: cast_nullable_to_non_nullable
as DoneeType?,studentLabel: freezed == studentLabel ? _self.studentLabel : studentLabel // ignore: cast_nullable_to_non_nullable
as StudentLabel?,tuitionLevel: freezed == tuitionLevel ? _self.tuitionLevel : tuitionLevel // ignore: cast_nullable_to_non_nullable
as TuitionLevel?,scholarshipCents: freezed == scholarshipCents ? _self.scholarshipCents : scholarshipCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,
  ));
}


}

// dart format on
