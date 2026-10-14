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

@JsonKey(name: 'regime_code') RegimeCode get regimeCode;@JsonKey(name: 'regime_name') String get regimeName; Treatment get treatment; RegimeEstimateStatus get status;@JsonKey(name: 'missing_inputs') List<MissingInput> get missingInputs;/// Plain-language text for a partial or informational regime (or an odd 621 month); `null` otherwise.
@JsonKey(includeIfNull: true) String? get explanation;@JsonKey(includeIfNull: true) TaxResult? get isr;@JsonKey(includeIfNull: true, name: 'iva_components') IvaComponents? get ivaComponents;/// ISR/IVA the taxpayer withheld from its suppliers in this regime (PM only).
@JsonKey(name: 'withheld_to_remit_cents') NonNegativeCents get withheldToRemitCents;/// A rule of this regime's computation is not validated by a CPA yet (⚠️ on its lines).
@JsonKey(name: 'cpa_review_pending') bool get cpaReviewPending;@JsonKey(name: 'cpa_topics') List<CpaTopic> get cpaTopics;@JsonKey(name: 'parameter_set') ParameterSetRef get parameterSet;
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
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RegimeEstimate&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.regimeName, _this.regimeName) || other.regimeName == _this.regimeName)&&(identical(other.treatment, _this.treatment) || other.treatment == _this.treatment)&&(identical(other.status, _this.status) || other.status == _this.status)&&const DeepCollectionEquality().equals(other.missingInputs, _this.missingInputs)&&(identical(other.explanation, _this.explanation) || other.explanation == _this.explanation)&&(identical(other.isr, _this.isr) || other.isr == _this.isr)&&(identical(other.ivaComponents, _this.ivaComponents) || other.ivaComponents == _this.ivaComponents)&&(identical(other.withheldToRemitCents, _this.withheldToRemitCents) || other.withheldToRemitCents == _this.withheldToRemitCents)&&(identical(other.cpaReviewPending, _this.cpaReviewPending) || other.cpaReviewPending == _this.cpaReviewPending)&&const DeepCollectionEquality().equals(other.cpaTopics, _this.cpaTopics)&&(identical(other.parameterSet, _this.parameterSet) || other.parameterSet == _this.parameterSet));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RegimeEstimate;
  return Object.hash(runtimeType,_this.regimeCode,_this.regimeName,_this.treatment,_this.status,const DeepCollectionEquality().hash(_this.missingInputs),_this.explanation,_this.isr,_this.ivaComponents,_this.withheldToRemitCents,_this.cpaReviewPending,const DeepCollectionEquality().hash(_this.cpaTopics),_this.parameterSet);
}

@override
String toString() {
  final _this = this as RegimeEstimate;
  return 'RegimeEstimate(regimeCode: ${_this.regimeCode}, regimeName: ${_this.regimeName}, treatment: ${_this.treatment}, status: ${_this.status}, missingInputs: ${_this.missingInputs}, explanation: ${_this.explanation}, isr: ${_this.isr}, ivaComponents: ${_this.ivaComponents}, withheldToRemitCents: ${_this.withheldToRemitCents}, cpaReviewPending: ${_this.cpaReviewPending}, cpaTopics: ${_this.cpaTopics}, parameterSet: ${_this.parameterSet})';
}


}

/// @nodoc
abstract mixin class $RegimeEstimateCopyWith<$Res>  {
  factory $RegimeEstimateCopyWith(RegimeEstimate value, $Res Function(RegimeEstimate) _then) = _$RegimeEstimateCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'regime_name') String regimeName, Treatment treatment, RegimeEstimateStatus status,@JsonKey(name: 'missing_inputs') List<MissingInput> missingInputs,@JsonKey(includeIfNull: true) String? explanation,@JsonKey(includeIfNull: true) TaxResult? isr,@JsonKey(includeIfNull: true, name: 'iva_components') IvaComponents? ivaComponents,@JsonKey(name: 'withheld_to_remit_cents') NonNegativeCents withheldToRemitCents,@JsonKey(name: 'cpa_review_pending') bool cpaReviewPending,@JsonKey(name: 'cpa_topics') List<CpaTopic> cpaTopics,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet
});


$TaxResultCopyWith<$Res>? get isr;$IvaComponentsCopyWith<$Res>? get ivaComponents;$ParameterSetRefCopyWith<$Res> get parameterSet;

}
/// @nodoc
class _$RegimeEstimateCopyWithImpl<$Res>
    implements $RegimeEstimateCopyWith<$Res> {
  _$RegimeEstimateCopyWithImpl(this._self, this._then);

  final RegimeEstimate _self;
  final $Res Function(RegimeEstimate) _then;

/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = null,Object? regimeName = null,Object? treatment = null,Object? status = null,Object? missingInputs = null,Object? explanation = freezed,Object? isr = freezed,Object? ivaComponents = freezed,Object? withheldToRemitCents = null,Object? cpaReviewPending = null,Object? cpaTopics = null,Object? parameterSet = null,}) {
  return _then(RegimeEstimate(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,regimeName: null == regimeName ? _self.regimeName : regimeName // ignore: cast_nullable_to_non_nullable
as String,treatment: null == treatment ? _self.treatment : treatment // ignore: cast_nullable_to_non_nullable
as Treatment,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as RegimeEstimateStatus,missingInputs: null == missingInputs ? _self.missingInputs : missingInputs // ignore: cast_nullable_to_non_nullable
as List<MissingInput>,explanation: freezed == explanation ? _self.explanation : explanation // ignore: cast_nullable_to_non_nullable
as String?,isr: freezed == isr ? _self.isr : isr // ignore: cast_nullable_to_non_nullable
as TaxResult?,ivaComponents: freezed == ivaComponents ? _self.ivaComponents : ivaComponents // ignore: cast_nullable_to_non_nullable
as IvaComponents?,withheldToRemitCents: null == withheldToRemitCents ? _self.withheldToRemitCents : withheldToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,cpaReviewPending: null == cpaReviewPending ? _self.cpaReviewPending : cpaReviewPending // ignore: cast_nullable_to_non_nullable
as bool,cpaTopics: null == cpaTopics ? _self.cpaTopics : cpaTopics // ignore: cast_nullable_to_non_nullable
as List<CpaTopic>,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,
  ));
}
/// Create a copy of RegimeEstimate
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
}/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ParameterSetRefCopyWith<$Res> get parameterSet {
  
  return $ParameterSetRefCopyWith<$Res>(_self.parameterSet, (value) {
    return _then(_self.copyWith(parameterSet: value));
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'regime_name')  String regimeName,  Treatment treatment,  RegimeEstimateStatus status, @JsonKey(name: 'missing_inputs')  List<MissingInput> missingInputs, @JsonKey(includeIfNull: true)  String? explanation, @JsonKey(includeIfNull: true)  TaxResult? isr, @JsonKey(includeIfNull: true, name: 'iva_components')  IvaComponents? ivaComponents, @JsonKey(name: 'withheld_to_remit_cents')  NonNegativeCents withheldToRemitCents, @JsonKey(name: 'cpa_review_pending')  bool cpaReviewPending, @JsonKey(name: 'cpa_topics')  List<CpaTopic> cpaTopics, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RegimeEstimate() when $default != null:
return $default(_that.regimeCode,_that.regimeName,_that.treatment,_that.status,_that.missingInputs,_that.explanation,_that.isr,_that.ivaComponents,_that.withheldToRemitCents,_that.cpaReviewPending,_that.cpaTopics,_that.parameterSet);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'regime_name')  String regimeName,  Treatment treatment,  RegimeEstimateStatus status, @JsonKey(name: 'missing_inputs')  List<MissingInput> missingInputs, @JsonKey(includeIfNull: true)  String? explanation, @JsonKey(includeIfNull: true)  TaxResult? isr, @JsonKey(includeIfNull: true, name: 'iva_components')  IvaComponents? ivaComponents, @JsonKey(name: 'withheld_to_remit_cents')  NonNegativeCents withheldToRemitCents, @JsonKey(name: 'cpa_review_pending')  bool cpaReviewPending, @JsonKey(name: 'cpa_topics')  List<CpaTopic> cpaTopics, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet)  $default,) {final _that = this;
switch (_that) {
case _RegimeEstimate():
return $default(_that.regimeCode,_that.regimeName,_that.treatment,_that.status,_that.missingInputs,_that.explanation,_that.isr,_that.ivaComponents,_that.withheldToRemitCents,_that.cpaReviewPending,_that.cpaTopics,_that.parameterSet);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'regime_name')  String regimeName,  Treatment treatment,  RegimeEstimateStatus status, @JsonKey(name: 'missing_inputs')  List<MissingInput> missingInputs, @JsonKey(includeIfNull: true)  String? explanation, @JsonKey(includeIfNull: true)  TaxResult? isr, @JsonKey(includeIfNull: true, name: 'iva_components')  IvaComponents? ivaComponents, @JsonKey(name: 'withheld_to_remit_cents')  NonNegativeCents withheldToRemitCents, @JsonKey(name: 'cpa_review_pending')  bool cpaReviewPending, @JsonKey(name: 'cpa_topics')  List<CpaTopic> cpaTopics, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet)?  $default,) {final _that = this;
switch (_that) {
case _RegimeEstimate() when $default != null:
return $default(_that.regimeCode,_that.regimeName,_that.treatment,_that.status,_that.missingInputs,_that.explanation,_that.isr,_that.ivaComponents,_that.withheldToRemitCents,_that.cpaReviewPending,_that.cpaTopics,_that.parameterSet);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RegimeEstimate implements RegimeEstimate {
  const _RegimeEstimate({@JsonKey(name: 'regime_code') required this.regimeCode, @JsonKey(name: 'regime_name') required this.regimeName, required this.treatment, required this.status, @JsonKey(name: 'missing_inputs') required  List<MissingInput> missingInputs, @JsonKey(includeIfNull: true) required this.explanation, @JsonKey(includeIfNull: true) required this.isr, @JsonKey(includeIfNull: true, name: 'iva_components') required this.ivaComponents, @JsonKey(name: 'withheld_to_remit_cents') required this.withheldToRemitCents, @JsonKey(name: 'cpa_review_pending') required this.cpaReviewPending, @JsonKey(name: 'cpa_topics') required  List<CpaTopic> cpaTopics, @JsonKey(name: 'parameter_set') required this.parameterSet}): _missingInputs = missingInputs,_cpaTopics = cpaTopics;
  factory _RegimeEstimate.fromJson(Map<String, dynamic> json) => _$RegimeEstimateFromJson(json);

@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
@override@JsonKey(name: 'regime_name') final  String regimeName;
@override final  Treatment treatment;
@override final  RegimeEstimateStatus status;
 final  List<MissingInput> _missingInputs;
@override@JsonKey(name: 'missing_inputs') List<MissingInput> get missingInputs {
  if (_missingInputs is EqualUnmodifiableListView) return _missingInputs;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_missingInputs);
}

/// Plain-language text for a partial or informational regime (or an odd 621 month); `null` otherwise.
@override@JsonKey(includeIfNull: true) final  String? explanation;
@override@JsonKey(includeIfNull: true) final  TaxResult? isr;
@override@JsonKey(includeIfNull: true, name: 'iva_components') final  IvaComponents? ivaComponents;
/// ISR/IVA the taxpayer withheld from its suppliers in this regime (PM only).
@override@JsonKey(name: 'withheld_to_remit_cents') final  NonNegativeCents withheldToRemitCents;
/// A rule of this regime's computation is not validated by a CPA yet (⚠️ on its lines).
@override@JsonKey(name: 'cpa_review_pending') final  bool cpaReviewPending;
 final  List<CpaTopic> _cpaTopics;
@override@JsonKey(name: 'cpa_topics') List<CpaTopic> get cpaTopics {
  if (_cpaTopics is EqualUnmodifiableListView) return _cpaTopics;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_cpaTopics);
}

@override@JsonKey(name: 'parameter_set') final  ParameterSetRef parameterSet;

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
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RegimeEstimate&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.regimeName, regimeName) || other.regimeName == regimeName)&&(identical(other.treatment, treatment) || other.treatment == treatment)&&(identical(other.status, status) || other.status == status)&&const DeepCollectionEquality().equals(other.missingInputs, _missingInputs)&&(identical(other.explanation, explanation) || other.explanation == explanation)&&(identical(other.isr, isr) || other.isr == isr)&&(identical(other.ivaComponents, ivaComponents) || other.ivaComponents == ivaComponents)&&(identical(other.withheldToRemitCents, withheldToRemitCents) || other.withheldToRemitCents == withheldToRemitCents)&&(identical(other.cpaReviewPending, cpaReviewPending) || other.cpaReviewPending == cpaReviewPending)&&const DeepCollectionEquality().equals(other.cpaTopics, _cpaTopics)&&(identical(other.parameterSet, parameterSet) || other.parameterSet == parameterSet));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,regimeName,treatment,status,const DeepCollectionEquality().hash(_missingInputs),explanation,isr,ivaComponents,withheldToRemitCents,cpaReviewPending,const DeepCollectionEquality().hash(_cpaTopics),parameterSet);
}

@override
String toString() {
    return 'RegimeEstimate(regimeCode: $regimeCode, regimeName: $regimeName, treatment: $treatment, status: $status, missingInputs: $missingInputs, explanation: $explanation, isr: $isr, ivaComponents: $ivaComponents, withheldToRemitCents: $withheldToRemitCents, cpaReviewPending: $cpaReviewPending, cpaTopics: $cpaTopics, parameterSet: $parameterSet)';
}


}

/// @nodoc
abstract mixin class _$RegimeEstimateCopyWith<$Res> implements $RegimeEstimateCopyWith<$Res> {
  factory _$RegimeEstimateCopyWith(_RegimeEstimate value, $Res Function(_RegimeEstimate) _then) = __$RegimeEstimateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'regime_name') String regimeName, Treatment treatment, RegimeEstimateStatus status,@JsonKey(name: 'missing_inputs') List<MissingInput> missingInputs,@JsonKey(includeIfNull: true) String? explanation,@JsonKey(includeIfNull: true) TaxResult? isr,@JsonKey(includeIfNull: true, name: 'iva_components') IvaComponents? ivaComponents,@JsonKey(name: 'withheld_to_remit_cents') NonNegativeCents withheldToRemitCents,@JsonKey(name: 'cpa_review_pending') bool cpaReviewPending,@JsonKey(name: 'cpa_topics') List<CpaTopic> cpaTopics,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet
});


@override $TaxResultCopyWith<$Res>? get isr;@override $IvaComponentsCopyWith<$Res>? get ivaComponents;@override $ParameterSetRefCopyWith<$Res> get parameterSet;

}
/// @nodoc
class __$RegimeEstimateCopyWithImpl<$Res>
    implements _$RegimeEstimateCopyWith<$Res> {
  __$RegimeEstimateCopyWithImpl(this._self, this._then);

  final _RegimeEstimate _self;
  final $Res Function(_RegimeEstimate) _then;

/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = null,Object? regimeName = null,Object? treatment = null,Object? status = null,Object? missingInputs = null,Object? explanation = freezed,Object? isr = freezed,Object? ivaComponents = freezed,Object? withheldToRemitCents = null,Object? cpaReviewPending = null,Object? cpaTopics = null,Object? parameterSet = null,}) {
  return _then(_RegimeEstimate(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,regimeName: null == regimeName ? _self.regimeName : regimeName // ignore: cast_nullable_to_non_nullable
as String,treatment: null == treatment ? _self.treatment : treatment // ignore: cast_nullable_to_non_nullable
as Treatment,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as RegimeEstimateStatus,missingInputs: null == missingInputs ? _self._missingInputs : missingInputs // ignore: cast_nullable_to_non_nullable
as List<MissingInput>,explanation: freezed == explanation ? _self.explanation : explanation // ignore: cast_nullable_to_non_nullable
as String?,isr: freezed == isr ? _self.isr : isr // ignore: cast_nullable_to_non_nullable
as TaxResult?,ivaComponents: freezed == ivaComponents ? _self.ivaComponents : ivaComponents // ignore: cast_nullable_to_non_nullable
as IvaComponents?,withheldToRemitCents: null == withheldToRemitCents ? _self.withheldToRemitCents : withheldToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,cpaReviewPending: null == cpaReviewPending ? _self.cpaReviewPending : cpaReviewPending // ignore: cast_nullable_to_non_nullable
as bool,cpaTopics: null == cpaTopics ? _self._cpaTopics : cpaTopics // ignore: cast_nullable_to_non_nullable
as List<CpaTopic>,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,
  ));
}

/// Create a copy of RegimeEstimate
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
}/// Create a copy of RegimeEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ParameterSetRefCopyWith<$Res> get parameterSet {
  
  return $ParameterSetRefCopyWith<$Res>(_self.parameterSet, (value) {
    return _then(_self.copyWith(parameterSet: value));
  });
}
}

// dart format on
