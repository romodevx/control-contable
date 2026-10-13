// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'expense_create.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ExpenseCreate {

@JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? get regimeCode;@JsonKey(includeIfNull: false, name: 'category_code') String? get categoryCode;@JsonKey(includeIfNull: false) String? get description;/// Subtotal before IVA.
@JsonKey(includeIfNull: false, name: 'amount_cents') NonNegativeCents? get amountCents;@JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? get ivaRate;@JsonKey(includeIfNull: false, name: 'iva_cents') NonNegativeCents? get ivaCents;@JsonKey(includeIfNull: false, name: 'payment_method') PaymentMethod? get paymentMethod;@JsonKey(includeIfNull: false, name: 'has_cfdi') bool? get hasCfdi;@JsonKey(includeIfNull: false, name: 'paid_on') DateTime? get paidOn;/// ISR the taxpayer (typically a PM) withheld from the supplier; must be remitted.
@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? get isrWithheldCents;/// IVA the taxpayer withheld from the supplier (LIVA 1-A); remitted in full, creditable when paid.
@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? get ivaWithheldCents;@JsonKey(includeIfNull: false) Uuid? get id;
/// Create a copy of ExpenseCreate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ExpenseCreateCopyWith<ExpenseCreate> get copyWith => _$ExpenseCreateCopyWithImpl<ExpenseCreate>(this as ExpenseCreate, _$identity);

  /// Serializes this ExpenseCreate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ExpenseCreate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ExpenseCreate&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.categoryCode, _this.categoryCode) || other.categoryCode == _this.categoryCode)&&(identical(other.description, _this.description) || other.description == _this.description)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.ivaRate, _this.ivaRate) || other.ivaRate == _this.ivaRate)&&(identical(other.ivaCents, _this.ivaCents) || other.ivaCents == _this.ivaCents)&&(identical(other.paymentMethod, _this.paymentMethod) || other.paymentMethod == _this.paymentMethod)&&(identical(other.hasCfdi, _this.hasCfdi) || other.hasCfdi == _this.hasCfdi)&&(identical(other.paidOn, _this.paidOn) || other.paidOn == _this.paidOn)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.ivaWithheldCents, _this.ivaWithheldCents) || other.ivaWithheldCents == _this.ivaWithheldCents)&&(identical(other.id, _this.id) || other.id == _this.id));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ExpenseCreate;
  return Object.hash(runtimeType,_this.regimeCode,_this.categoryCode,_this.description,_this.amountCents,_this.ivaRate,_this.ivaCents,_this.paymentMethod,_this.hasCfdi,_this.paidOn,_this.isrWithheldCents,_this.ivaWithheldCents,_this.id);
}

@override
String toString() {
  final _this = this as ExpenseCreate;
  return 'ExpenseCreate(regimeCode: ${_this.regimeCode}, categoryCode: ${_this.categoryCode}, description: ${_this.description}, amountCents: ${_this.amountCents}, ivaRate: ${_this.ivaRate}, ivaCents: ${_this.ivaCents}, paymentMethod: ${_this.paymentMethod}, hasCfdi: ${_this.hasCfdi}, paidOn: ${_this.paidOn}, isrWithheldCents: ${_this.isrWithheldCents}, ivaWithheldCents: ${_this.ivaWithheldCents}, id: ${_this.id})';
}


}

/// @nodoc
abstract mixin class $ExpenseCreateCopyWith<$Res>  {
  factory $ExpenseCreateCopyWith(ExpenseCreate value, $Res Function(ExpenseCreate) _then) = _$ExpenseCreateCopyWithImpl;
@useResult
$Res call({
@JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? regimeCode,@JsonKey(includeIfNull: false, name: 'category_code') String? categoryCode,@JsonKey(includeIfNull: false) String? description,@JsonKey(includeIfNull: false, name: 'amount_cents') NonNegativeCents? amountCents,@JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? ivaRate,@JsonKey(includeIfNull: false, name: 'iva_cents') NonNegativeCents? ivaCents,@JsonKey(includeIfNull: false, name: 'payment_method') PaymentMethod? paymentMethod,@JsonKey(includeIfNull: false, name: 'has_cfdi') bool? hasCfdi,@JsonKey(includeIfNull: false, name: 'paid_on') DateTime? paidOn,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,@JsonKey(includeIfNull: false) Uuid? id
});




}
/// @nodoc
class _$ExpenseCreateCopyWithImpl<$Res>
    implements $ExpenseCreateCopyWith<$Res> {
  _$ExpenseCreateCopyWithImpl(this._self, this._then);

  final ExpenseCreate _self;
  final $Res Function(ExpenseCreate) _then;

/// Create a copy of ExpenseCreate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = freezed,Object? categoryCode = freezed,Object? description = freezed,Object? amountCents = freezed,Object? ivaRate = freezed,Object? ivaCents = freezed,Object? paymentMethod = freezed,Object? hasCfdi = freezed,Object? paidOn = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,Object? id = freezed,}) {
  return _then(ExpenseCreate(
regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode?,categoryCode: freezed == categoryCode ? _self.categoryCode : categoryCode // ignore: cast_nullable_to_non_nullable
as String?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as String?,amountCents: freezed == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaRate: freezed == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate?,ivaCents: freezed == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,paymentMethod: freezed == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as PaymentMethod?,hasCfdi: freezed == hasCfdi ? _self.hasCfdi : hasCfdi // ignore: cast_nullable_to_non_nullable
as bool?,paidOn: freezed == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as DateTime?,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaWithheldCents: freezed == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,
  ));
}

}


/// Adds pattern-matching-related methods to [ExpenseCreate].
extension ExpenseCreatePatterns on ExpenseCreate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ExpenseCreate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ExpenseCreate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ExpenseCreate value)  $default,){
final _that = this;
switch (_that) {
case _ExpenseCreate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ExpenseCreate value)?  $default,){
final _that = this;
switch (_that) {
case _ExpenseCreate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: false, name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(includeIfNull: false, name: 'category_code')  String? categoryCode, @JsonKey(includeIfNull: false)  String? description, @JsonKey(includeIfNull: false, name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(includeIfNull: false, name: 'payment_method')  PaymentMethod? paymentMethod, @JsonKey(includeIfNull: false, name: 'has_cfdi')  bool? hasCfdi, @JsonKey(includeIfNull: false, name: 'paid_on')  DateTime? paidOn, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false)  Uuid? id)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ExpenseCreate() when $default != null:
return $default(_that.regimeCode,_that.categoryCode,_that.description,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.hasCfdi,_that.paidOn,_that.isrWithheldCents,_that.ivaWithheldCents,_that.id);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: false, name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(includeIfNull: false, name: 'category_code')  String? categoryCode, @JsonKey(includeIfNull: false)  String? description, @JsonKey(includeIfNull: false, name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(includeIfNull: false, name: 'payment_method')  PaymentMethod? paymentMethod, @JsonKey(includeIfNull: false, name: 'has_cfdi')  bool? hasCfdi, @JsonKey(includeIfNull: false, name: 'paid_on')  DateTime? paidOn, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false)  Uuid? id)  $default,) {final _that = this;
switch (_that) {
case _ExpenseCreate():
return $default(_that.regimeCode,_that.categoryCode,_that.description,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.hasCfdi,_that.paidOn,_that.isrWithheldCents,_that.ivaWithheldCents,_that.id);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(includeIfNull: false, name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(includeIfNull: false, name: 'category_code')  String? categoryCode, @JsonKey(includeIfNull: false)  String? description, @JsonKey(includeIfNull: false, name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(includeIfNull: false, name: 'payment_method')  PaymentMethod? paymentMethod, @JsonKey(includeIfNull: false, name: 'has_cfdi')  bool? hasCfdi, @JsonKey(includeIfNull: false, name: 'paid_on')  DateTime? paidOn, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false)  Uuid? id)?  $default,) {final _that = this;
switch (_that) {
case _ExpenseCreate() when $default != null:
return $default(_that.regimeCode,_that.categoryCode,_that.description,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.hasCfdi,_that.paidOn,_that.isrWithheldCents,_that.ivaWithheldCents,_that.id);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ExpenseCreate implements ExpenseCreate {
  const _ExpenseCreate({@JsonKey(includeIfNull: false, name: 'regime_code') this.regimeCode, @JsonKey(includeIfNull: false, name: 'category_code') this.categoryCode, @JsonKey(includeIfNull: false) this.description, @JsonKey(includeIfNull: false, name: 'amount_cents') this.amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate') this.ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents') this.ivaCents, @JsonKey(includeIfNull: false, name: 'payment_method') this.paymentMethod, @JsonKey(includeIfNull: false, name: 'has_cfdi') this.hasCfdi, @JsonKey(includeIfNull: false, name: 'paid_on') this.paidOn, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents') this.isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents') this.ivaWithheldCents, @JsonKey(includeIfNull: false) this.id});
  factory _ExpenseCreate.fromJson(Map<String, dynamic> json) => _$ExpenseCreateFromJson(json);

@override@JsonKey(includeIfNull: false, name: 'regime_code') final  RegimeCode? regimeCode;
@override@JsonKey(includeIfNull: false, name: 'category_code') final  String? categoryCode;
@override@JsonKey(includeIfNull: false) final  String? description;
/// Subtotal before IVA.
@override@JsonKey(includeIfNull: false, name: 'amount_cents') final  NonNegativeCents? amountCents;
@override@JsonKey(includeIfNull: false, name: 'iva_rate') final  IvaRate? ivaRate;
@override@JsonKey(includeIfNull: false, name: 'iva_cents') final  NonNegativeCents? ivaCents;
@override@JsonKey(includeIfNull: false, name: 'payment_method') final  PaymentMethod? paymentMethod;
@override@JsonKey(includeIfNull: false, name: 'has_cfdi') final  bool? hasCfdi;
@override@JsonKey(includeIfNull: false, name: 'paid_on') final  DateTime? paidOn;
/// ISR the taxpayer (typically a PM) withheld from the supplier; must be remitted.
@override@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') final  NonNegativeCents? isrWithheldCents;
/// IVA the taxpayer withheld from the supplier (LIVA 1-A); remitted in full, creditable when paid.
@override@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') final  NonNegativeCents? ivaWithheldCents;
@override@JsonKey(includeIfNull: false) final  Uuid? id;

/// Create a copy of ExpenseCreate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ExpenseCreateCopyWith<_ExpenseCreate> get copyWith => __$ExpenseCreateCopyWithImpl<_ExpenseCreate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ExpenseCreateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ExpenseCreate&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.categoryCode, categoryCode) || other.categoryCode == categoryCode)&&(identical(other.description, description) || other.description == description)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.ivaRate, ivaRate) || other.ivaRate == ivaRate)&&(identical(other.ivaCents, ivaCents) || other.ivaCents == ivaCents)&&(identical(other.paymentMethod, paymentMethod) || other.paymentMethod == paymentMethod)&&(identical(other.hasCfdi, hasCfdi) || other.hasCfdi == hasCfdi)&&(identical(other.paidOn, paidOn) || other.paidOn == paidOn)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.ivaWithheldCents, ivaWithheldCents) || other.ivaWithheldCents == ivaWithheldCents)&&(identical(other.id, id) || other.id == id));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,categoryCode,description,amountCents,ivaRate,ivaCents,paymentMethod,hasCfdi,paidOn,isrWithheldCents,ivaWithheldCents,id);
}

@override
String toString() {
    return 'ExpenseCreate(regimeCode: $regimeCode, categoryCode: $categoryCode, description: $description, amountCents: $amountCents, ivaRate: $ivaRate, ivaCents: $ivaCents, paymentMethod: $paymentMethod, hasCfdi: $hasCfdi, paidOn: $paidOn, isrWithheldCents: $isrWithheldCents, ivaWithheldCents: $ivaWithheldCents, id: $id)';
}


}

/// @nodoc
abstract mixin class _$ExpenseCreateCopyWith<$Res> implements $ExpenseCreateCopyWith<$Res> {
  factory _$ExpenseCreateCopyWith(_ExpenseCreate value, $Res Function(_ExpenseCreate) _then) = __$ExpenseCreateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? regimeCode,@JsonKey(includeIfNull: false, name: 'category_code') String? categoryCode,@JsonKey(includeIfNull: false) String? description,@JsonKey(includeIfNull: false, name: 'amount_cents') NonNegativeCents? amountCents,@JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? ivaRate,@JsonKey(includeIfNull: false, name: 'iva_cents') NonNegativeCents? ivaCents,@JsonKey(includeIfNull: false, name: 'payment_method') PaymentMethod? paymentMethod,@JsonKey(includeIfNull: false, name: 'has_cfdi') bool? hasCfdi,@JsonKey(includeIfNull: false, name: 'paid_on') DateTime? paidOn,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,@JsonKey(includeIfNull: false) Uuid? id
});




}
/// @nodoc
class __$ExpenseCreateCopyWithImpl<$Res>
    implements _$ExpenseCreateCopyWith<$Res> {
  __$ExpenseCreateCopyWithImpl(this._self, this._then);

  final _ExpenseCreate _self;
  final $Res Function(_ExpenseCreate) _then;

/// Create a copy of ExpenseCreate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = freezed,Object? categoryCode = freezed,Object? description = freezed,Object? amountCents = freezed,Object? ivaRate = freezed,Object? ivaCents = freezed,Object? paymentMethod = freezed,Object? hasCfdi = freezed,Object? paidOn = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,Object? id = freezed,}) {
  return _then(_ExpenseCreate(
regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode?,categoryCode: freezed == categoryCode ? _self.categoryCode : categoryCode // ignore: cast_nullable_to_non_nullable
as String?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as String?,amountCents: freezed == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaRate: freezed == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate?,ivaCents: freezed == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,paymentMethod: freezed == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as PaymentMethod?,hasCfdi: freezed == hasCfdi ? _self.hasCfdi : hasCfdi // ignore: cast_nullable_to_non_nullable
as bool?,paidOn: freezed == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as DateTime?,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaWithheldCents: freezed == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,
  ));
}


}

// dart format on
