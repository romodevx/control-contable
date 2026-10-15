// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_section.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualSection {

 AnnualSectionKind get kind; List<RegimeCode> get regimes; AnnualSectionStatus get status;@JsonKey(name: 'missing_inputs') List<AnnualMissingInput> get missingInputs;@JsonKey(includeIfNull: true) AnnualObligation? get obligation;@JsonKey(name: 'obligation_reasons') List<String> get obligationReasons;@JsonKey(includeIfNull: true) String? get deadline;@JsonKey(includeIfNull: true) AnnualTax? get tax;@JsonKey(includeIfNull: true, name: 'pm_extras') PmAnnualExtras? get pmExtras;@JsonKey(includeIfNull: true, name: 'resico_check') ResicoPfCheck? get resicoCheck;@JsonKey(includeIfNull: true) String? get explanation; List<AnnualNote> get notes;/// A rule of this section is not validated by a CPA yet (⚠️, strong banner).
@JsonKey(name: 'cpa_review_pending') bool get cpaReviewPending;@JsonKey(name: 'cpa_topics') List<CpaTopic> get cpaTopics;@JsonKey(name: 'parameter_set') ParameterSetRef get parameterSet;
/// Create a copy of AnnualSection
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualSectionCopyWith<AnnualSection> get copyWith => _$AnnualSectionCopyWithImpl<AnnualSection>(this as AnnualSection, _$identity);

  /// Serializes this AnnualSection to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualSection;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualSection&&(identical(other.kind, _this.kind) || other.kind == _this.kind)&&const DeepCollectionEquality().equals(other.regimes, _this.regimes)&&(identical(other.status, _this.status) || other.status == _this.status)&&const DeepCollectionEquality().equals(other.missingInputs, _this.missingInputs)&&(identical(other.obligation, _this.obligation) || other.obligation == _this.obligation)&&const DeepCollectionEquality().equals(other.obligationReasons, _this.obligationReasons)&&(identical(other.deadline, _this.deadline) || other.deadline == _this.deadline)&&(identical(other.tax, _this.tax) || other.tax == _this.tax)&&(identical(other.pmExtras, _this.pmExtras) || other.pmExtras == _this.pmExtras)&&(identical(other.resicoCheck, _this.resicoCheck) || other.resicoCheck == _this.resicoCheck)&&(identical(other.explanation, _this.explanation) || other.explanation == _this.explanation)&&const DeepCollectionEquality().equals(other.notes, _this.notes)&&(identical(other.cpaReviewPending, _this.cpaReviewPending) || other.cpaReviewPending == _this.cpaReviewPending)&&const DeepCollectionEquality().equals(other.cpaTopics, _this.cpaTopics)&&(identical(other.parameterSet, _this.parameterSet) || other.parameterSet == _this.parameterSet));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualSection;
  return Object.hash(runtimeType,_this.kind,const DeepCollectionEquality().hash(_this.regimes),_this.status,const DeepCollectionEquality().hash(_this.missingInputs),_this.obligation,const DeepCollectionEquality().hash(_this.obligationReasons),_this.deadline,_this.tax,_this.pmExtras,_this.resicoCheck,_this.explanation,const DeepCollectionEquality().hash(_this.notes),_this.cpaReviewPending,const DeepCollectionEquality().hash(_this.cpaTopics),_this.parameterSet);
}

@override
String toString() {
  final _this = this as AnnualSection;
  return 'AnnualSection(kind: ${_this.kind}, regimes: ${_this.regimes}, status: ${_this.status}, missingInputs: ${_this.missingInputs}, obligation: ${_this.obligation}, obligationReasons: ${_this.obligationReasons}, deadline: ${_this.deadline}, tax: ${_this.tax}, pmExtras: ${_this.pmExtras}, resicoCheck: ${_this.resicoCheck}, explanation: ${_this.explanation}, notes: ${_this.notes}, cpaReviewPending: ${_this.cpaReviewPending}, cpaTopics: ${_this.cpaTopics}, parameterSet: ${_this.parameterSet})';
}


}

/// @nodoc
abstract mixin class $AnnualSectionCopyWith<$Res>  {
  factory $AnnualSectionCopyWith(AnnualSection value, $Res Function(AnnualSection) _then) = _$AnnualSectionCopyWithImpl;
@useResult
$Res call({
 AnnualSectionKind kind, List<RegimeCode> regimes, AnnualSectionStatus status,@JsonKey(name: 'missing_inputs') List<AnnualMissingInput> missingInputs,@JsonKey(includeIfNull: true) AnnualObligation? obligation,@JsonKey(name: 'obligation_reasons') List<String> obligationReasons,@JsonKey(includeIfNull: true) String? deadline,@JsonKey(includeIfNull: true) AnnualTax? tax,@JsonKey(includeIfNull: true, name: 'pm_extras') PmAnnualExtras? pmExtras,@JsonKey(includeIfNull: true, name: 'resico_check') ResicoPfCheck? resicoCheck,@JsonKey(includeIfNull: true) String? explanation, List<AnnualNote> notes,@JsonKey(name: 'cpa_review_pending') bool cpaReviewPending,@JsonKey(name: 'cpa_topics') List<CpaTopic> cpaTopics,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet
});


$AnnualTaxCopyWith<$Res>? get tax;$PmAnnualExtrasCopyWith<$Res>? get pmExtras;$ResicoPfCheckCopyWith<$Res>? get resicoCheck;$ParameterSetRefCopyWith<$Res> get parameterSet;

}
/// @nodoc
class _$AnnualSectionCopyWithImpl<$Res>
    implements $AnnualSectionCopyWith<$Res> {
  _$AnnualSectionCopyWithImpl(this._self, this._then);

  final AnnualSection _self;
  final $Res Function(AnnualSection) _then;

/// Create a copy of AnnualSection
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? kind = null,Object? regimes = null,Object? status = null,Object? missingInputs = null,Object? obligation = freezed,Object? obligationReasons = null,Object? deadline = freezed,Object? tax = freezed,Object? pmExtras = freezed,Object? resicoCheck = freezed,Object? explanation = freezed,Object? notes = null,Object? cpaReviewPending = null,Object? cpaTopics = null,Object? parameterSet = null,}) {
  return _then(AnnualSection(
kind: null == kind ? _self.kind : kind // ignore: cast_nullable_to_non_nullable
as AnnualSectionKind,regimes: null == regimes ? _self.regimes : regimes // ignore: cast_nullable_to_non_nullable
as List<RegimeCode>,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as AnnualSectionStatus,missingInputs: null == missingInputs ? _self.missingInputs : missingInputs // ignore: cast_nullable_to_non_nullable
as List<AnnualMissingInput>,obligation: freezed == obligation ? _self.obligation : obligation // ignore: cast_nullable_to_non_nullable
as AnnualObligation?,obligationReasons: null == obligationReasons ? _self.obligationReasons : obligationReasons // ignore: cast_nullable_to_non_nullable
as List<String>,deadline: freezed == deadline ? _self.deadline : deadline // ignore: cast_nullable_to_non_nullable
as String?,tax: freezed == tax ? _self.tax : tax // ignore: cast_nullable_to_non_nullable
as AnnualTax?,pmExtras: freezed == pmExtras ? _self.pmExtras : pmExtras // ignore: cast_nullable_to_non_nullable
as PmAnnualExtras?,resicoCheck: freezed == resicoCheck ? _self.resicoCheck : resicoCheck // ignore: cast_nullable_to_non_nullable
as ResicoPfCheck?,explanation: freezed == explanation ? _self.explanation : explanation // ignore: cast_nullable_to_non_nullable
as String?,notes: null == notes ? _self.notes : notes // ignore: cast_nullable_to_non_nullable
as List<AnnualNote>,cpaReviewPending: null == cpaReviewPending ? _self.cpaReviewPending : cpaReviewPending // ignore: cast_nullable_to_non_nullable
as bool,cpaTopics: null == cpaTopics ? _self.cpaTopics : cpaTopics // ignore: cast_nullable_to_non_nullable
as List<CpaTopic>,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,
  ));
}
/// Create a copy of AnnualSection
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AnnualTaxCopyWith<$Res>? get tax {
    if (_self.tax == null) {
    return null;
  }

  return $AnnualTaxCopyWith<$Res>(_self.tax!, (value) {
    return _then(_self.copyWith(tax: value));
  });
}/// Create a copy of AnnualSection
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$PmAnnualExtrasCopyWith<$Res>? get pmExtras {
    if (_self.pmExtras == null) {
    return null;
  }

  return $PmAnnualExtrasCopyWith<$Res>(_self.pmExtras!, (value) {
    return _then(_self.copyWith(pmExtras: value));
  });
}/// Create a copy of AnnualSection
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ResicoPfCheckCopyWith<$Res>? get resicoCheck {
    if (_self.resicoCheck == null) {
    return null;
  }

  return $ResicoPfCheckCopyWith<$Res>(_self.resicoCheck!, (value) {
    return _then(_self.copyWith(resicoCheck: value));
  });
}/// Create a copy of AnnualSection
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ParameterSetRefCopyWith<$Res> get parameterSet {
  
  return $ParameterSetRefCopyWith<$Res>(_self.parameterSet, (value) {
    return _then(_self.copyWith(parameterSet: value));
  });
}
}


/// Adds pattern-matching-related methods to [AnnualSection].
extension AnnualSectionPatterns on AnnualSection {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualSection value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualSection() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualSection value)  $default,){
final _that = this;
switch (_that) {
case _AnnualSection():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualSection value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualSection() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( AnnualSectionKind kind,  List<RegimeCode> regimes,  AnnualSectionStatus status, @JsonKey(name: 'missing_inputs')  List<AnnualMissingInput> missingInputs, @JsonKey(includeIfNull: true)  AnnualObligation? obligation, @JsonKey(name: 'obligation_reasons')  List<String> obligationReasons, @JsonKey(includeIfNull: true)  String? deadline, @JsonKey(includeIfNull: true)  AnnualTax? tax, @JsonKey(includeIfNull: true, name: 'pm_extras')  PmAnnualExtras? pmExtras, @JsonKey(includeIfNull: true, name: 'resico_check')  ResicoPfCheck? resicoCheck, @JsonKey(includeIfNull: true)  String? explanation,  List<AnnualNote> notes, @JsonKey(name: 'cpa_review_pending')  bool cpaReviewPending, @JsonKey(name: 'cpa_topics')  List<CpaTopic> cpaTopics, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualSection() when $default != null:
return $default(_that.kind,_that.regimes,_that.status,_that.missingInputs,_that.obligation,_that.obligationReasons,_that.deadline,_that.tax,_that.pmExtras,_that.resicoCheck,_that.explanation,_that.notes,_that.cpaReviewPending,_that.cpaTopics,_that.parameterSet);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( AnnualSectionKind kind,  List<RegimeCode> regimes,  AnnualSectionStatus status, @JsonKey(name: 'missing_inputs')  List<AnnualMissingInput> missingInputs, @JsonKey(includeIfNull: true)  AnnualObligation? obligation, @JsonKey(name: 'obligation_reasons')  List<String> obligationReasons, @JsonKey(includeIfNull: true)  String? deadline, @JsonKey(includeIfNull: true)  AnnualTax? tax, @JsonKey(includeIfNull: true, name: 'pm_extras')  PmAnnualExtras? pmExtras, @JsonKey(includeIfNull: true, name: 'resico_check')  ResicoPfCheck? resicoCheck, @JsonKey(includeIfNull: true)  String? explanation,  List<AnnualNote> notes, @JsonKey(name: 'cpa_review_pending')  bool cpaReviewPending, @JsonKey(name: 'cpa_topics')  List<CpaTopic> cpaTopics, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet)  $default,) {final _that = this;
switch (_that) {
case _AnnualSection():
return $default(_that.kind,_that.regimes,_that.status,_that.missingInputs,_that.obligation,_that.obligationReasons,_that.deadline,_that.tax,_that.pmExtras,_that.resicoCheck,_that.explanation,_that.notes,_that.cpaReviewPending,_that.cpaTopics,_that.parameterSet);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( AnnualSectionKind kind,  List<RegimeCode> regimes,  AnnualSectionStatus status, @JsonKey(name: 'missing_inputs')  List<AnnualMissingInput> missingInputs, @JsonKey(includeIfNull: true)  AnnualObligation? obligation, @JsonKey(name: 'obligation_reasons')  List<String> obligationReasons, @JsonKey(includeIfNull: true)  String? deadline, @JsonKey(includeIfNull: true)  AnnualTax? tax, @JsonKey(includeIfNull: true, name: 'pm_extras')  PmAnnualExtras? pmExtras, @JsonKey(includeIfNull: true, name: 'resico_check')  ResicoPfCheck? resicoCheck, @JsonKey(includeIfNull: true)  String? explanation,  List<AnnualNote> notes, @JsonKey(name: 'cpa_review_pending')  bool cpaReviewPending, @JsonKey(name: 'cpa_topics')  List<CpaTopic> cpaTopics, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet)?  $default,) {final _that = this;
switch (_that) {
case _AnnualSection() when $default != null:
return $default(_that.kind,_that.regimes,_that.status,_that.missingInputs,_that.obligation,_that.obligationReasons,_that.deadline,_that.tax,_that.pmExtras,_that.resicoCheck,_that.explanation,_that.notes,_that.cpaReviewPending,_that.cpaTopics,_that.parameterSet);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualSection implements AnnualSection {
  const _AnnualSection({required this.kind, required  List<RegimeCode> regimes, required this.status, @JsonKey(name: 'missing_inputs') required  List<AnnualMissingInput> missingInputs, @JsonKey(includeIfNull: true) required this.obligation, @JsonKey(name: 'obligation_reasons') required  List<String> obligationReasons, @JsonKey(includeIfNull: true) required this.deadline, @JsonKey(includeIfNull: true) required this.tax, @JsonKey(includeIfNull: true, name: 'pm_extras') required this.pmExtras, @JsonKey(includeIfNull: true, name: 'resico_check') required this.resicoCheck, @JsonKey(includeIfNull: true) required this.explanation, required  List<AnnualNote> notes, @JsonKey(name: 'cpa_review_pending') required this.cpaReviewPending, @JsonKey(name: 'cpa_topics') required  List<CpaTopic> cpaTopics, @JsonKey(name: 'parameter_set') required this.parameterSet}): _regimes = regimes,_missingInputs = missingInputs,_obligationReasons = obligationReasons,_notes = notes,_cpaTopics = cpaTopics;
  factory _AnnualSection.fromJson(Map<String, dynamic> json) => _$AnnualSectionFromJson(json);

@override final  AnnualSectionKind kind;
 final  List<RegimeCode> _regimes;
@override List<RegimeCode> get regimes {
  if (_regimes is EqualUnmodifiableListView) return _regimes;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_regimes);
}

@override final  AnnualSectionStatus status;
 final  List<AnnualMissingInput> _missingInputs;
@override@JsonKey(name: 'missing_inputs') List<AnnualMissingInput> get missingInputs {
  if (_missingInputs is EqualUnmodifiableListView) return _missingInputs;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_missingInputs);
}

@override@JsonKey(includeIfNull: true) final  AnnualObligation? obligation;
 final  List<String> _obligationReasons;
@override@JsonKey(name: 'obligation_reasons') List<String> get obligationReasons {
  if (_obligationReasons is EqualUnmodifiableListView) return _obligationReasons;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_obligationReasons);
}

@override@JsonKey(includeIfNull: true) final  String? deadline;
@override@JsonKey(includeIfNull: true) final  AnnualTax? tax;
@override@JsonKey(includeIfNull: true, name: 'pm_extras') final  PmAnnualExtras? pmExtras;
@override@JsonKey(includeIfNull: true, name: 'resico_check') final  ResicoPfCheck? resicoCheck;
@override@JsonKey(includeIfNull: true) final  String? explanation;
 final  List<AnnualNote> _notes;
@override List<AnnualNote> get notes {
  if (_notes is EqualUnmodifiableListView) return _notes;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_notes);
}

/// A rule of this section is not validated by a CPA yet (⚠️, strong banner).
@override@JsonKey(name: 'cpa_review_pending') final  bool cpaReviewPending;
 final  List<CpaTopic> _cpaTopics;
@override@JsonKey(name: 'cpa_topics') List<CpaTopic> get cpaTopics {
  if (_cpaTopics is EqualUnmodifiableListView) return _cpaTopics;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_cpaTopics);
}

@override@JsonKey(name: 'parameter_set') final  ParameterSetRef parameterSet;

/// Create a copy of AnnualSection
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualSectionCopyWith<_AnnualSection> get copyWith => __$AnnualSectionCopyWithImpl<_AnnualSection>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualSectionToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualSection&&(identical(other.kind, kind) || other.kind == kind)&&const DeepCollectionEquality().equals(other.regimes, _regimes)&&(identical(other.status, status) || other.status == status)&&const DeepCollectionEquality().equals(other.missingInputs, _missingInputs)&&(identical(other.obligation, obligation) || other.obligation == obligation)&&const DeepCollectionEquality().equals(other.obligationReasons, _obligationReasons)&&(identical(other.deadline, deadline) || other.deadline == deadline)&&(identical(other.tax, tax) || other.tax == tax)&&(identical(other.pmExtras, pmExtras) || other.pmExtras == pmExtras)&&(identical(other.resicoCheck, resicoCheck) || other.resicoCheck == resicoCheck)&&(identical(other.explanation, explanation) || other.explanation == explanation)&&const DeepCollectionEquality().equals(other.notes, _notes)&&(identical(other.cpaReviewPending, cpaReviewPending) || other.cpaReviewPending == cpaReviewPending)&&const DeepCollectionEquality().equals(other.cpaTopics, _cpaTopics)&&(identical(other.parameterSet, parameterSet) || other.parameterSet == parameterSet));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,kind,const DeepCollectionEquality().hash(_regimes),status,const DeepCollectionEquality().hash(_missingInputs),obligation,const DeepCollectionEquality().hash(_obligationReasons),deadline,tax,pmExtras,resicoCheck,explanation,const DeepCollectionEquality().hash(_notes),cpaReviewPending,const DeepCollectionEquality().hash(_cpaTopics),parameterSet);
}

@override
String toString() {
    return 'AnnualSection(kind: $kind, regimes: $regimes, status: $status, missingInputs: $missingInputs, obligation: $obligation, obligationReasons: $obligationReasons, deadline: $deadline, tax: $tax, pmExtras: $pmExtras, resicoCheck: $resicoCheck, explanation: $explanation, notes: $notes, cpaReviewPending: $cpaReviewPending, cpaTopics: $cpaTopics, parameterSet: $parameterSet)';
}


}

/// @nodoc
abstract mixin class _$AnnualSectionCopyWith<$Res> implements $AnnualSectionCopyWith<$Res> {
  factory _$AnnualSectionCopyWith(_AnnualSection value, $Res Function(_AnnualSection) _then) = __$AnnualSectionCopyWithImpl;
@override @useResult
$Res call({
 AnnualSectionKind kind, List<RegimeCode> regimes, AnnualSectionStatus status,@JsonKey(name: 'missing_inputs') List<AnnualMissingInput> missingInputs,@JsonKey(includeIfNull: true) AnnualObligation? obligation,@JsonKey(name: 'obligation_reasons') List<String> obligationReasons,@JsonKey(includeIfNull: true) String? deadline,@JsonKey(includeIfNull: true) AnnualTax? tax,@JsonKey(includeIfNull: true, name: 'pm_extras') PmAnnualExtras? pmExtras,@JsonKey(includeIfNull: true, name: 'resico_check') ResicoPfCheck? resicoCheck,@JsonKey(includeIfNull: true) String? explanation, List<AnnualNote> notes,@JsonKey(name: 'cpa_review_pending') bool cpaReviewPending,@JsonKey(name: 'cpa_topics') List<CpaTopic> cpaTopics,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet
});


@override $AnnualTaxCopyWith<$Res>? get tax;@override $PmAnnualExtrasCopyWith<$Res>? get pmExtras;@override $ResicoPfCheckCopyWith<$Res>? get resicoCheck;@override $ParameterSetRefCopyWith<$Res> get parameterSet;

}
/// @nodoc
class __$AnnualSectionCopyWithImpl<$Res>
    implements _$AnnualSectionCopyWith<$Res> {
  __$AnnualSectionCopyWithImpl(this._self, this._then);

  final _AnnualSection _self;
  final $Res Function(_AnnualSection) _then;

/// Create a copy of AnnualSection
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? kind = null,Object? regimes = null,Object? status = null,Object? missingInputs = null,Object? obligation = freezed,Object? obligationReasons = null,Object? deadline = freezed,Object? tax = freezed,Object? pmExtras = freezed,Object? resicoCheck = freezed,Object? explanation = freezed,Object? notes = null,Object? cpaReviewPending = null,Object? cpaTopics = null,Object? parameterSet = null,}) {
  return _then(_AnnualSection(
kind: null == kind ? _self.kind : kind // ignore: cast_nullable_to_non_nullable
as AnnualSectionKind,regimes: null == regimes ? _self._regimes : regimes // ignore: cast_nullable_to_non_nullable
as List<RegimeCode>,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as AnnualSectionStatus,missingInputs: null == missingInputs ? _self._missingInputs : missingInputs // ignore: cast_nullable_to_non_nullable
as List<AnnualMissingInput>,obligation: freezed == obligation ? _self.obligation : obligation // ignore: cast_nullable_to_non_nullable
as AnnualObligation?,obligationReasons: null == obligationReasons ? _self._obligationReasons : obligationReasons // ignore: cast_nullable_to_non_nullable
as List<String>,deadline: freezed == deadline ? _self.deadline : deadline // ignore: cast_nullable_to_non_nullable
as String?,tax: freezed == tax ? _self.tax : tax // ignore: cast_nullable_to_non_nullable
as AnnualTax?,pmExtras: freezed == pmExtras ? _self.pmExtras : pmExtras // ignore: cast_nullable_to_non_nullable
as PmAnnualExtras?,resicoCheck: freezed == resicoCheck ? _self.resicoCheck : resicoCheck // ignore: cast_nullable_to_non_nullable
as ResicoPfCheck?,explanation: freezed == explanation ? _self.explanation : explanation // ignore: cast_nullable_to_non_nullable
as String?,notes: null == notes ? _self._notes : notes // ignore: cast_nullable_to_non_nullable
as List<AnnualNote>,cpaReviewPending: null == cpaReviewPending ? _self.cpaReviewPending : cpaReviewPending // ignore: cast_nullable_to_non_nullable
as bool,cpaTopics: null == cpaTopics ? _self._cpaTopics : cpaTopics // ignore: cast_nullable_to_non_nullable
as List<CpaTopic>,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,
  ));
}

/// Create a copy of AnnualSection
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AnnualTaxCopyWith<$Res>? get tax {
    if (_self.tax == null) {
    return null;
  }

  return $AnnualTaxCopyWith<$Res>(_self.tax!, (value) {
    return _then(_self.copyWith(tax: value));
  });
}/// Create a copy of AnnualSection
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$PmAnnualExtrasCopyWith<$Res>? get pmExtras {
    if (_self.pmExtras == null) {
    return null;
  }

  return $PmAnnualExtrasCopyWith<$Res>(_self.pmExtras!, (value) {
    return _then(_self.copyWith(pmExtras: value));
  });
}/// Create a copy of AnnualSection
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ResicoPfCheckCopyWith<$Res>? get resicoCheck {
    if (_self.resicoCheck == null) {
    return null;
  }

  return $ResicoPfCheckCopyWith<$Res>(_self.resicoCheck!, (value) {
    return _then(_self.copyWith(resicoCheck: value));
  });
}/// Create a copy of AnnualSection
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
