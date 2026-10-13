// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'regime_estimate.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RegimeEstimate {

@JsonKey(name: 'regime_code') RegimeCode get regimeCode;@JsonKey(name: 'regime_name') String get regimeName; Treatment get treatment; RegimeEstimateStatus get status;@JsonKey(name: 'parameter_set') ParameterSetRef get parameterSet;@JsonKey(name: 'missing_inputs') List<String>? get missingInputs; String? get explanation; TaxResult? get isr;@JsonKey(name: 'iva_components') IvaComponents? get ivaComponents;/// ISR/IVA the taxpayer withheld from suppliers in this regime.
@JsonKey(name: 'withheld_to_remit_cents') NonNegativeCents? get withheldToRemitCents;
/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RegimeEstimateCopyWith<RegimeEstimate> get copyWith => _$RegimeEstimateCopyWithImpl<RegimeEstimate>(this as RegimeEstimate, _$identity);

  /// Serializes this RegimeEstimate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RegimeEstimate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RegimeEstimate&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.regimeName, _this.regimeName) || other.regimeName == _this.regimeName)&&(identical(other.treatment, _this.treatment) || other.treatment == _this.treatment)&&(identical(other.status, _this.status) || other.status == _this.status)&&(identical(other.parameterSet, _this.parameterSet) || other.parameterSet == _this.parameterSet)&&const DeepCollectionEquality().equals(other.missingInputs, _this.missingInputs)&&(identical(other.explanation, _this.explanation) || other.explanation == _this.explanation)&&(identical(other.isr, _this.isr) || other.isr == _this.isr)&&(identical(other.ivaComponents, _this.ivaComponents) || other.ivaComponents == _this.ivaComponents)&&(identical(other.withheldToRemitCents, _this.withheldToRemitCents) || other.withheldToRemitCents == _this.withheldToRemitCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RegimeEstimate;
  return Object.hash(runtimeType,_this.regimeCode,_this.regimeName,_this.treatment,_this.status,_this.parameterSet,const DeepCollectionEquality().hash(_this.missingInputs),_this.explanation,_this.isr,_this.ivaComponents,_this.withheldToRemitCents);
}

@override
String toString() {
  final _this = this as RegimeEstimate;
  return 'RegimeEstimate(regimeCode: ${_this.regimeCode}, regimeName: ${_this.regimeName}, treatment: ${_this.treatment}, status: ${_this.status}, parameterSet: ${_this.parameterSet}, missingInputs: ${_this.missingInputs}, explanation: ${_this.explanation}, isr: ${_this.isr}, ivaComponents: ${_this.ivaComponents}, withheldToRemitCents: ${_this.withheldToRemitCents})';
}


}

/// @nodoc
abstract mixin class $RegimeEstimateCopyWith<$Res>  {
  factory $RegimeEstimateCopyWith(RegimeEstimate value, $Res Function(RegimeEstimate) _then) = _$RegimeEstimateCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'regime_name') String regimeName, Treatment treatment, RegimeEstimateStatus status,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet,@JsonKey(name: 'missing_inputs') List<String>? missingInputs, String? explanation, TaxResult? isr,@JsonKey(name: 'iva_components') IvaComponents? ivaComponents,@JsonKey(name: 'withheld_to_remit_cents') NonNegativeCents? withheldToRemitCents
});


$ParameterSetRefCopyWith<$Res> get parameterSet;$TaxResultCopyWith<$Res>? get isr;$IvaComponentsCopyWith<$Res>? get ivaComponents;

}
/// @nodoc
class _$RegimeEstimateCopyWithImpl<$Res>
    implements $RegimeEstimateCopyWith<$Res> {
  _$RegimeEstimateCopyWithImpl(this._self, this._then);

  final RegimeEstimate _self;
  final $Res Function(RegimeEstimate) _then;

/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = null,Object? regimeName = null,Object? treatment = null,Object? status = null,Object? parameterSet = null,Object? missingInputs = freezed,Object? explanation = freezed,Object? isr = freezed,Object? ivaComponents = freezed,Object? withheldToRemitCents = freezed,}) {
  return _then(RegimeEstimate(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,regimeName: null == regimeName ? _self.regimeName : regimeName // ignore: cast_nullable_to_non_nullable
as String,treatment: null == treatment ? _self.treatment : treatment // ignore: cast_nullable_to_non_nullable
as Treatment,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as RegimeEstimateStatus,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,missingInputs: freezed == missingInputs ? _self.missingInputs : missingInputs // ignore: cast_nullable_to_non_nullable
as List<String>?,explanation: freezed == explanation ? _self.explanation : explanation // ignore: cast_nullable_to_non_nullable
as String?,isr: freezed == isr ? _self.isr : isr // ignore: cast_nullable_to_non_nullable
as TaxResult?,ivaComponents: freezed == ivaComponents ? _self.ivaComponents : ivaComponents // ignore: cast_nullable_to_non_nullable
as IvaComponents?,withheldToRemitCents: freezed == withheldToRemitCents ? _self.withheldToRemitCents : withheldToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,
  ));
}
/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ParameterSetRefCopyWith<$Res> get parameterSet {
  
  return $ParameterSetRefCopyWith<$Res>(_self.parameterSet, (value) {
    return _then(_self.copyWith(parameterSet: value));
  });
}/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TaxResultCopyWith<$Res>? get isr {
    if (_self.isr == null) {
    return null;
  }

  return $TaxResultCopyWith<$Res>(_self.isr!, (value) {
    return _then(_self.copyWith(isr: value));
  });
}/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$IvaComponentsCopyWith<$Res>? get ivaComponents {
    if (_self.ivaComponents == null) {
    return null;
  }

  return $IvaComponentsCopyWith<$Res>(_self.ivaComponents!, (value) {
    return _then(_self.copyWith(ivaComponents: value));
  });
}
}


/// Adds pattern-matching-related methods to [RegimeEstimate].
extension RegimeEstimatePatterns on RegimeEstimate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RegimeEstimate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RegimeEstimate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RegimeEstimate value)  $default,){
final _that = this;
switch (_that) {
case _RegimeEstimate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RegimeEstimate value)?  $default,){
final _that = this;
switch (_that) {
case _RegimeEstimate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'regime_name')  String regimeName,  Treatment treatment,  RegimeEstimateStatus status, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(name: 'missing_inputs')  List<String>? missingInputs,  String? explanation,  TaxResult? isr, @JsonKey(name: 'iva_components')  IvaComponents? ivaComponents, @JsonKey(name: 'withheld_to_remit_cents')  NonNegativeCents? withheldToRemitCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RegimeEstimate() when $default != null:
return $default(_that.regimeCode,_that.regimeName,_that.treatment,_that.status,_that.parameterSet,_that.missingInputs,_that.explanation,_that.isr,_that.ivaComponents,_that.withheldToRemitCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'regime_name')  String regimeName,  Treatment treatment,  RegimeEstimateStatus status, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(name: 'missing_inputs')  List<String>? missingInputs,  String? explanation,  TaxResult? isr, @JsonKey(name: 'iva_components')  IvaComponents? ivaComponents, @JsonKey(name: 'withheld_to_remit_cents')  NonNegativeCents? withheldToRemitCents)  $default,) {final _that = this;
switch (_that) {
case _RegimeEstimate():
return $default(_that.regimeCode,_that.regimeName,_that.treatment,_that.status,_that.parameterSet,_that.missingInputs,_that.explanation,_that.isr,_that.ivaComponents,_that.withheldToRemitCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'regime_name')  String regimeName,  Treatment treatment,  RegimeEstimateStatus status, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(name: 'missing_inputs')  List<String>? missingInputs,  String? explanation,  TaxResult? isr, @JsonKey(name: 'iva_components')  IvaComponents? ivaComponents, @JsonKey(name: 'withheld_to_remit_cents')  NonNegativeCents? withheldToRemitCents)?  $default,) {final _that = this;
switch (_that) {
case _RegimeEstimate() when $default != null:
return $default(_that.regimeCode,_that.regimeName,_that.treatment,_that.status,_that.parameterSet,_that.missingInputs,_that.explanation,_that.isr,_that.ivaComponents,_that.withheldToRemitCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RegimeEstimate implements RegimeEstimate {
  const _RegimeEstimate({@JsonKey(name: 'regime_code') required this.regimeCode, @JsonKey(name: 'regime_name') required this.regimeName, required this.treatment, required this.status, @JsonKey(name: 'parameter_set') required this.parameterSet, @JsonKey(name: 'missing_inputs')  List<String>? missingInputs, this.explanation, this.isr, @JsonKey(name: 'iva_components') this.ivaComponents, @JsonKey(name: 'withheld_to_remit_cents') this.withheldToRemitCents}): _missingInputs = missingInputs;
  factory _RegimeEstimate.fromJson(Map<String, dynamic> json) => _$RegimeEstimateFromJson(json);

@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
@override@JsonKey(name: 'regime_name') final  String regimeName;
@override final  Treatment treatment;
@override final  RegimeEstimateStatus status;
@override@JsonKey(name: 'parameter_set') final  ParameterSetRef parameterSet;
 final  List<String>? _missingInputs;
@override@JsonKey(name: 'missing_inputs') List<String>? get missingInputs {
  final value = _missingInputs;
  if (value == null) return null;
  if (_missingInputs is EqualUnmodifiableListView) return _missingInputs;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(value);
}

@override final  String? explanation;
@override final  TaxResult? isr;
@override@JsonKey(name: 'iva_components') final  IvaComponents? ivaComponents;
/// ISR/IVA the taxpayer withheld from suppliers in this regime.
@override@JsonKey(name: 'withheld_to_remit_cents') final  NonNegativeCents? withheldToRemitCents;

/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RegimeEstimateCopyWith<_RegimeEstimate> get copyWith => __$RegimeEstimateCopyWithImpl<_RegimeEstimate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RegimeEstimateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RegimeEstimate&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.regimeName, regimeName) || other.regimeName == regimeName)&&(identical(other.treatment, treatment) || other.treatment == treatment)&&(identical(other.status, status) || other.status == status)&&(identical(other.parameterSet, parameterSet) || other.parameterSet == parameterSet)&&const DeepCollectionEquality().equals(other.missingInputs, _missingInputs)&&(identical(other.explanation, explanation) || other.explanation == explanation)&&(identical(other.isr, isr) || other.isr == isr)&&(identical(other.ivaComponents, ivaComponents) || other.ivaComponents == ivaComponents)&&(identical(other.withheldToRemitCents, withheldToRemitCents) || other.withheldToRemitCents == withheldToRemitCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,regimeName,treatment,status,parameterSet,const DeepCollectionEquality().hash(_missingInputs),explanation,isr,ivaComponents,withheldToRemitCents);
}

@override
String toString() {
    return 'RegimeEstimate(regimeCode: $regimeCode, regimeName: $regimeName, treatment: $treatment, status: $status, parameterSet: $parameterSet, missingInputs: $missingInputs, explanation: $explanation, isr: $isr, ivaComponents: $ivaComponents, withheldToRemitCents: $withheldToRemitCents)';
}


}

/// @nodoc
abstract mixin class _$RegimeEstimateCopyWith<$Res> implements $RegimeEstimateCopyWith<$Res> {
  factory _$RegimeEstimateCopyWith(_RegimeEstimate value, $Res Function(_RegimeEstimate) _then) = __$RegimeEstimateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'regime_name') String regimeName, Treatment treatment, RegimeEstimateStatus status,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet,@JsonKey(name: 'missing_inputs') List<String>? missingInputs, String? explanation, TaxResult? isr,@JsonKey(name: 'iva_components') IvaComponents? ivaComponents,@JsonKey(name: 'withheld_to_remit_cents') NonNegativeCents? withheldToRemitCents
});


@override $ParameterSetRefCopyWith<$Res> get parameterSet;@override $TaxResultCopyWith<$Res>? get isr;@override $IvaComponentsCopyWith<$Res>? get ivaComponents;

}
/// @nodoc
class __$RegimeEstimateCopyWithImpl<$Res>
    implements _$RegimeEstimateCopyWith<$Res> {
  __$RegimeEstimateCopyWithImpl(this._self, this._then);

  final _RegimeEstimate _self;
  final $Res Function(_RegimeEstimate) _then;

/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = null,Object? regimeName = null,Object? treatment = null,Object? status = null,Object? parameterSet = null,Object? missingInputs = freezed,Object? explanation = freezed,Object? isr = freezed,Object? ivaComponents = freezed,Object? withheldToRemitCents = freezed,}) {
  return _then(_RegimeEstimate(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,regimeName: null == regimeName ? _self.regimeName : regimeName // ignore: cast_nullable_to_non_nullable
as String,treatment: null == treatment ? _self.treatment : treatment // ignore: cast_nullable_to_non_nullable
as Treatment,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as RegimeEstimateStatus,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,missingInputs: freezed == missingInputs ? _self._missingInputs : missingInputs // ignore: cast_nullable_to_non_nullable
as List<String>?,explanation: freezed == explanation ? _self.explanation : explanation // ignore: cast_nullable_to_non_nullable
as String?,isr: freezed == isr ? _self.isr : isr // ignore: cast_nullable_to_non_nullable
as TaxResult?,ivaComponents: freezed == ivaComponents ? _self.ivaComponents : ivaComponents // ignore: cast_nullable_to_non_nullable
as IvaComponents?,withheldToRemitCents: freezed == withheldToRemitCents ? _self.withheldToRemitCents : withheldToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,
  ));
}

/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ParameterSetRefCopyWith<$Res> get parameterSet {
  
  return $ParameterSetRefCopyWith<$Res>(_self.parameterSet, (value) {
    return _then(_self.copyWith(parameterSet: value));
  });
}/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TaxResultCopyWith<$Res>? get isr {
    if (_self.isr == null) {
    return null;
  }

  return $TaxResultCopyWith<$Res>(_self.isr!, (value) {
    return _then(_self.copyWith(isr: value));
  });
}/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$IvaComponentsCopyWith<$Res>? get ivaComponents {
    if (_self.ivaComponents == null) {
    return null;
  }

  return $IvaComponentsCopyWith<$Res>(_self.ivaComponents!, (value) {
    return _then(_self.copyWith(ivaComponents: value));
  });
}
}

// dart format on
