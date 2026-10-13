// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'expense.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Expense {

 Uuid get id;@JsonKey(name: 'taxpayer_id') Uuid get taxpayerId; List<ExpenseWarning> get warnings; Legend get legend;@JsonKey(name: 'created_at') DateTime get createdAt;@JsonKey(name: 'updated_at') DateTime get updatedAt;@JsonKey(name: 'regime_code') RegimeCode? get regimeCode;@JsonKey(name: 'category_code') String? get categoryCode; String? get description;/// Subtotal before IVA.
@JsonKey(name: 'amount_cents') NonNegativeCents? get amountCents;@JsonKey(name: 'iva_rate') IvaRate? get ivaRate;@JsonKey(name: 'iva_cents') NonNegativeCents? get ivaCents;@JsonKey(name: 'payment_method') PaymentMethod? get paymentMethod;@JsonKey(name: 'has_cfdi') bool? get hasCfdi;@JsonKey(name: 'paid_on') DateTime? get paidOn;/// ISR the taxpayer (typically a PM) withheld from the supplier; must be remitted.
@JsonKey(name: 'isr_withheld_cents') NonNegativeCents? get isrWithheldCents;/// IVA the taxpayer withheld from the supplier (LIVA 1-A); remitted in full, creditable when paid.
@JsonKey(name: 'iva_withheld_cents') NonNegativeCents? get ivaWithheldCents;
/// Create a copy of Expense
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ExpenseCopyWith<Expense> get copyWith => _$ExpenseCopyWithImpl<Expense>(this as Expense, _$identity);

  /// Serializes this Expense to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Expense;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Expense&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.taxpayerId, _this.taxpayerId) || other.taxpayerId == _this.taxpayerId)&&const DeepCollectionEquality().equals(other.warnings, _this.warnings)&&(identical(other.legend, _this.legend) || other.legend == _this.legend)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.updatedAt, _this.updatedAt) || other.updatedAt == _this.updatedAt)&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.categoryCode, _this.categoryCode) || other.categoryCode == _this.categoryCode)&&(identical(other.description, _this.description) || other.description == _this.description)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.ivaRate, _this.ivaRate) || other.ivaRate == _this.ivaRate)&&(identical(other.ivaCents, _this.ivaCents) || other.ivaCents == _this.ivaCents)&&(identical(other.paymentMethod, _this.paymentMethod) || other.paymentMethod == _this.paymentMethod)&&(identical(other.hasCfdi, _this.hasCfdi) || other.hasCfdi == _this.hasCfdi)&&(identical(other.paidOn, _this.paidOn) || other.paidOn == _this.paidOn)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.ivaWithheldCents, _this.ivaWithheldCents) || other.ivaWithheldCents == _this.ivaWithheldCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Expense;
  return Object.hash(runtimeType,_this.id,_this.taxpayerId,const DeepCollectionEquality().hash(_this.warnings),_this.legend,_this.createdAt,_this.updatedAt,_this.regimeCode,_this.categoryCode,_this.description,_this.amountCents,_this.ivaRate,_this.ivaCents,_this.paymentMethod,_this.hasCfdi,_this.paidOn,_this.isrWithheldCents,_this.ivaWithheldCents);
}

@override
String toString() {
  final _this = this as Expense;
  return 'Expense(id: ${_this.id}, taxpayerId: ${_this.taxpayerId}, warnings: ${_this.warnings}, legend: ${_this.legend}, createdAt: ${_this.createdAt}, updatedAt: ${_this.updatedAt}, regimeCode: ${_this.regimeCode}, categoryCode: ${_this.categoryCode}, description: ${_this.description}, amountCents: ${_this.amountCents}, ivaRate: ${_this.ivaRate}, ivaCents: ${_this.ivaCents}, paymentMethod: ${_this.paymentMethod}, hasCfdi: ${_this.hasCfdi}, paidOn: ${_this.paidOn}, isrWithheldCents: ${_this.isrWithheldCents}, ivaWithheldCents: ${_this.ivaWithheldCents})';
}


}

/// @nodoc
abstract mixin class $ExpenseCopyWith<$Res>  {
  factory $ExpenseCopyWith(Expense value, $Res Function(Expense) _then) = _$ExpenseCopyWithImpl;
@useResult
$Res call({
 Uuid id,@JsonKey(name: 'taxpayer_id') Uuid taxpayerId, List<ExpenseWarning> warnings, Legend legend,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt,@JsonKey(name: 'regime_code') RegimeCode? regimeCode,@JsonKey(name: 'category_code') String? categoryCode, String? description,@JsonKey(name: 'amount_cents') NonNegativeCents? amountCents,@JsonKey(name: 'iva_rate') IvaRate? ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents? ivaCents,@JsonKey(name: 'payment_method') PaymentMethod? paymentMethod,@JsonKey(name: 'has_cfdi') bool? hasCfdi,@JsonKey(name: 'paid_on') DateTime? paidOn,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents
});




}
/// @nodoc
class _$ExpenseCopyWithImpl<$Res>
    implements $ExpenseCopyWith<$Res> {
  _$ExpenseCopyWithImpl(this._self, this._then);

  final Expense _self;
  final $Res Function(Expense) _then;

/// Create a copy of Expense
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? taxpayerId = null,Object? warnings = null,Object? legend = null,Object? createdAt = null,Object? updatedAt = null,Object? regimeCode = freezed,Object? categoryCode = freezed,Object? description = freezed,Object? amountCents = freezed,Object? ivaRate = freezed,Object? ivaCents = freezed,Object? paymentMethod = freezed,Object? hasCfdi = freezed,Object? paidOn = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,}) {
  return _then(Expense(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,warnings: null == warnings ? _self.warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<ExpenseWarning>,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
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
as NonNegativeCents?,
  ));
}

}


/// Adds pattern-matching-related methods to [Expense].
extension ExpensePatterns on Expense {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Expense value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Expense() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Expense value)  $default,){
final _that = this;
switch (_that) {
case _Expense():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Expense value)?  $default,){
final _that = this;
switch (_that) {
case _Expense() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  List<ExpenseWarning> warnings,  Legend legend, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt, @JsonKey(name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(name: 'category_code')  String? categoryCode,  String? description, @JsonKey(name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(name: 'payment_method')  PaymentMethod? paymentMethod, @JsonKey(name: 'has_cfdi')  bool? hasCfdi, @JsonKey(name: 'paid_on')  DateTime? paidOn, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Expense() when $default != null:
return $default(_that.id,_that.taxpayerId,_that.warnings,_that.legend,_that.createdAt,_that.updatedAt,_that.regimeCode,_that.categoryCode,_that.description,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.hasCfdi,_that.paidOn,_that.isrWithheldCents,_that.ivaWithheldCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  List<ExpenseWarning> warnings,  Legend legend, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt, @JsonKey(name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(name: 'category_code')  String? categoryCode,  String? description, @JsonKey(name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(name: 'payment_method')  PaymentMethod? paymentMethod, @JsonKey(name: 'has_cfdi')  bool? hasCfdi, @JsonKey(name: 'paid_on')  DateTime? paidOn, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents)  $default,) {final _that = this;
switch (_that) {
case _Expense():
return $default(_that.id,_that.taxpayerId,_that.warnings,_that.legend,_that.createdAt,_that.updatedAt,_that.regimeCode,_that.categoryCode,_that.description,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.hasCfdi,_that.paidOn,_that.isrWithheldCents,_that.ivaWithheldCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  List<ExpenseWarning> warnings,  Legend legend, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt, @JsonKey(name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(name: 'category_code')  String? categoryCode,  String? description, @JsonKey(name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(name: 'payment_method')  PaymentMethod? paymentMethod, @JsonKey(name: 'has_cfdi')  bool? hasCfdi, @JsonKey(name: 'paid_on')  DateTime? paidOn, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents)?  $default,) {final _that = this;
switch (_that) {
case _Expense() when $default != null:
return $default(_that.id,_that.taxpayerId,_that.warnings,_that.legend,_that.createdAt,_that.updatedAt,_that.regimeCode,_that.categoryCode,_that.description,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.hasCfdi,_that.paidOn,_that.isrWithheldCents,_that.ivaWithheldCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Expense implements Expense {
  const _Expense({required this.id, @JsonKey(name: 'taxpayer_id') required this.taxpayerId, required  List<ExpenseWarning> warnings, required this.legend, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(name: 'updated_at') required this.updatedAt, @JsonKey(name: 'regime_code') this.regimeCode, @JsonKey(name: 'category_code') this.categoryCode, this.description, @JsonKey(name: 'amount_cents') this.amountCents, @JsonKey(name: 'iva_rate') this.ivaRate, @JsonKey(name: 'iva_cents') this.ivaCents, @JsonKey(name: 'payment_method') this.paymentMethod, @JsonKey(name: 'has_cfdi') this.hasCfdi, @JsonKey(name: 'paid_on') this.paidOn, @JsonKey(name: 'isr_withheld_cents') this.isrWithheldCents, @JsonKey(name: 'iva_withheld_cents') this.ivaWithheldCents}): _warnings = warnings;
  factory _Expense.fromJson(Map<String, dynamic> json) => _$ExpenseFromJson(json);

@override final  Uuid id;
@override@JsonKey(name: 'taxpayer_id') final  Uuid taxpayerId;
 final  List<ExpenseWarning> _warnings;
@override List<ExpenseWarning> get warnings {
  if (_warnings is EqualUnmodifiableListView) return _warnings;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_warnings);
}

@override final  Legend legend;
@override@JsonKey(name: 'created_at') final  DateTime createdAt;
@override@JsonKey(name: 'updated_at') final  DateTime updatedAt;
@override@JsonKey(name: 'regime_code') final  RegimeCode? regimeCode;
@override@JsonKey(name: 'category_code') final  String? categoryCode;
@override final  String? description;
/// Subtotal before IVA.
@override@JsonKey(name: 'amount_cents') final  NonNegativeCents? amountCents;
@override@JsonKey(name: 'iva_rate') final  IvaRate? ivaRate;
@override@JsonKey(name: 'iva_cents') final  NonNegativeCents? ivaCents;
@override@JsonKey(name: 'payment_method') final  PaymentMethod? paymentMethod;
@override@JsonKey(name: 'has_cfdi') final  bool? hasCfdi;
@override@JsonKey(name: 'paid_on') final  DateTime? paidOn;
/// ISR the taxpayer (typically a PM) withheld from the supplier; must be remitted.
@override@JsonKey(name: 'isr_withheld_cents') final  NonNegativeCents? isrWithheldCents;
/// IVA the taxpayer withheld from the supplier (LIVA 1-A); remitted in full, creditable when paid.
@override@JsonKey(name: 'iva_withheld_cents') final  NonNegativeCents? ivaWithheldCents;

/// Create a copy of Expense
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ExpenseCopyWith<_Expense> get copyWith => __$ExpenseCopyWithImpl<_Expense>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ExpenseToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Expense&&(identical(other.id, id) || other.id == id)&&(identical(other.taxpayerId, taxpayerId) || other.taxpayerId == taxpayerId)&&const DeepCollectionEquality().equals(other.warnings, _warnings)&&(identical(other.legend, legend) || other.legend == legend)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt)&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.categoryCode, categoryCode) || other.categoryCode == categoryCode)&&(identical(other.description, description) || other.description == description)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.ivaRate, ivaRate) || other.ivaRate == ivaRate)&&(identical(other.ivaCents, ivaCents) || other.ivaCents == ivaCents)&&(identical(other.paymentMethod, paymentMethod) || other.paymentMethod == paymentMethod)&&(identical(other.hasCfdi, hasCfdi) || other.hasCfdi == hasCfdi)&&(identical(other.paidOn, paidOn) || other.paidOn == paidOn)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.ivaWithheldCents, ivaWithheldCents) || other.ivaWithheldCents == ivaWithheldCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,taxpayerId,const DeepCollectionEquality().hash(_warnings),legend,createdAt,updatedAt,regimeCode,categoryCode,description,amountCents,ivaRate,ivaCents,paymentMethod,hasCfdi,paidOn,isrWithheldCents,ivaWithheldCents);
}

@override
String toString() {
    return 'Expense(id: $id, taxpayerId: $taxpayerId, warnings: $warnings, legend: $legend, createdAt: $createdAt, updatedAt: $updatedAt, regimeCode: $regimeCode, categoryCode: $categoryCode, description: $description, amountCents: $amountCents, ivaRate: $ivaRate, ivaCents: $ivaCents, paymentMethod: $paymentMethod, hasCfdi: $hasCfdi, paidOn: $paidOn, isrWithheldCents: $isrWithheldCents, ivaWithheldCents: $ivaWithheldCents)';
}


}

/// @nodoc
abstract mixin class _$ExpenseCopyWith<$Res> implements $ExpenseCopyWith<$Res> {
  factory _$ExpenseCopyWith(_Expense value, $Res Function(_Expense) _then) = __$ExpenseCopyWithImpl;
@override @useResult
$Res call({
 Uuid id,@JsonKey(name: 'taxpayer_id') Uuid taxpayerId, List<ExpenseWarning> warnings, Legend legend,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt,@JsonKey(name: 'regime_code') RegimeCode? regimeCode,@JsonKey(name: 'category_code') String? categoryCode, String? description,@JsonKey(name: 'amount_cents') NonNegativeCents? amountCents,@JsonKey(name: 'iva_rate') IvaRate? ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents? ivaCents,@JsonKey(name: 'payment_method') PaymentMethod? paymentMethod,@JsonKey(name: 'has_cfdi') bool? hasCfdi,@JsonKey(name: 'paid_on') DateTime? paidOn,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents
});




}
/// @nodoc
class __$ExpenseCopyWithImpl<$Res>
    implements _$ExpenseCopyWith<$Res> {
  __$ExpenseCopyWithImpl(this._self, this._then);

  final _Expense _self;
  final $Res Function(_Expense) _then;

/// Create a copy of Expense
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? taxpayerId = null,Object? warnings = null,Object? legend = null,Object? createdAt = null,Object? updatedAt = null,Object? regimeCode = freezed,Object? categoryCode = freezed,Object? description = freezed,Object? amountCents = freezed,Object? ivaRate = freezed,Object? ivaCents = freezed,Object? paymentMethod = freezed,Object? hasCfdi = freezed,Object? paidOn = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,}) {
  return _then(_Expense(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,warnings: null == warnings ? _self._warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<ExpenseWarning>,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
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
as NonNegativeCents?,
  ));
}


}

// dart format on
