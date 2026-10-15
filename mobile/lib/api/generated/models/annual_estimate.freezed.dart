// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_estimate.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualEstimate {

@JsonKey(name: 'taxpayer_id') Uuid get taxpayerId;@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear;@JsonKey(name: 'person_type') PersonType get personType; AnnualLegend get legend; List<Disclaimer> get disclaimers;@JsonKey(name: 'year_status') AnnualYearStatus get yearStatus;@JsonKey(includeIfNull: true, name: 'data_through') Period? get dataThrough;@JsonKey(includeIfNull: true) String? get deadline;/// The sections whose `tax.balance` is the headline, each shown on its own (never netted).
@JsonKey(name: 'headline_sections') List<AnnualComputedSectionKind> get headlineSections; List<AnnualSection> get sections;@JsonKey(name: 'informative_income_over_500k') bool get informativeIncomeOver500k;@JsonKey(name: 'parameter_set') ParameterSetRef get parameterSet;@JsonKey(name: 'engine_version') EngineVersion get engineVersion;
/// Create a copy of AnnualEstimate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualEstimateCopyWith<AnnualEstimate> get copyWith => _$AnnualEstimateCopyWithImpl<AnnualEstimate>(this as AnnualEstimate, _$identity);

  /// Serializes this AnnualEstimate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualEstimate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualEstimate&&(identical(other.taxpayerId, _this.taxpayerId) || other.taxpayerId == _this.taxpayerId)&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&(identical(other.personType, _this.personType) || other.personType == _this.personType)&&(identical(other.legend, _this.legend) || other.legend == _this.legend)&&const DeepCollectionEquality().equals(other.disclaimers, _this.disclaimers)&&(identical(other.yearStatus, _this.yearStatus) || other.yearStatus == _this.yearStatus)&&(identical(other.dataThrough, _this.dataThrough) || other.dataThrough == _this.dataThrough)&&(identical(other.deadline, _this.deadline) || other.deadline == _this.deadline)&&const DeepCollectionEquality().equals(other.headlineSections, _this.headlineSections)&&const DeepCollectionEquality().equals(other.sections, _this.sections)&&(identical(other.informativeIncomeOver500k, _this.informativeIncomeOver500k) || other.informativeIncomeOver500k == _this.informativeIncomeOver500k)&&(identical(other.parameterSet, _this.parameterSet) || other.parameterSet == _this.parameterSet)&&(identical(other.engineVersion, _this.engineVersion) || other.engineVersion == _this.engineVersion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualEstimate;
  return Object.hash(runtimeType,_this.taxpayerId,_this.fiscalYear,_this.personType,_this.legend,const DeepCollectionEquality().hash(_this.disclaimers),_this.yearStatus,_this.dataThrough,_this.deadline,const DeepCollectionEquality().hash(_this.headlineSections),const DeepCollectionEquality().hash(_this.sections),_this.informativeIncomeOver500k,_this.parameterSet,_this.engineVersion);
}

@override
String toString() {
  final _this = this as AnnualEstimate;
  return 'AnnualEstimate(taxpayerId: ${_this.taxpayerId}, fiscalYear: ${_this.fiscalYear}, personType: ${_this.personType}, legend: ${_this.legend}, disclaimers: ${_this.disclaimers}, yearStatus: ${_this.yearStatus}, dataThrough: ${_this.dataThrough}, deadline: ${_this.deadline}, headlineSections: ${_this.headlineSections}, sections: ${_this.sections}, informativeIncomeOver500k: ${_this.informativeIncomeOver500k}, parameterSet: ${_this.parameterSet}, engineVersion: ${_this.engineVersion})';
}


}

/// @nodoc
abstract mixin class $AnnualEstimateCopyWith<$Res>  {
  factory $AnnualEstimateCopyWith(AnnualEstimate value, $Res Function(AnnualEstimate) _then) = _$AnnualEstimateCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear,@JsonKey(name: 'person_type') PersonType personType, AnnualLegend legend, List<Disclaimer> disclaimers,@JsonKey(name: 'year_status') AnnualYearStatus yearStatus,@JsonKey(includeIfNull: true, name: 'data_through') Period? dataThrough,@JsonKey(includeIfNull: true) String? deadline,@JsonKey(name: 'headline_sections') List<AnnualComputedSectionKind> headlineSections, List<AnnualSection> sections,@JsonKey(name: 'informative_income_over_500k') bool informativeIncomeOver500k,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet,@JsonKey(name: 'engine_version') EngineVersion engineVersion
});


$ParameterSetRefCopyWith<$Res> get parameterSet;

}
/// @nodoc
class _$AnnualEstimateCopyWithImpl<$Res>
    implements $AnnualEstimateCopyWith<$Res> {
  _$AnnualEstimateCopyWithImpl(this._self, this._then);

  final AnnualEstimate _self;
  final $Res Function(AnnualEstimate) _then;

/// Create a copy of AnnualEstimate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? taxpayerId = null,Object? fiscalYear = null,Object? personType = null,Object? legend = null,Object? disclaimers = null,Object? yearStatus = null,Object? dataThrough = freezed,Object? deadline = freezed,Object? headlineSections = null,Object? sections = null,Object? informativeIncomeOver500k = null,Object? parameterSet = null,Object? engineVersion = null,}) {
  return _then(AnnualEstimate(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,personType: null == personType ? _self.personType : personType // ignore: cast_nullable_to_non_nullable
as PersonType,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as AnnualLegend,disclaimers: null == disclaimers ? _self.disclaimers : disclaimers // ignore: cast_nullable_to_non_nullable
as List<Disclaimer>,yearStatus: null == yearStatus ? _self.yearStatus : yearStatus // ignore: cast_nullable_to_non_nullable
as AnnualYearStatus,dataThrough: freezed == dataThrough ? _self.dataThrough : dataThrough // ignore: cast_nullable_to_non_nullable
as Period?,deadline: freezed == deadline ? _self.deadline : deadline // ignore: cast_nullable_to_non_nullable
as String?,headlineSections: null == headlineSections ? _self.headlineSections : headlineSections // ignore: cast_nullable_to_non_nullable
as List<AnnualComputedSectionKind>,sections: null == sections ? _self.sections : sections // ignore: cast_nullable_to_non_nullable
as List<AnnualSection>,informativeIncomeOver500k: null == informativeIncomeOver500k ? _self.informativeIncomeOver500k : informativeIncomeOver500k // ignore: cast_nullable_to_non_nullable
as bool,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as EngineVersion,
  ));
}
/// Create a copy of AnnualEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ParameterSetRefCopyWith<$Res> get parameterSet {
  
  return $ParameterSetRefCopyWith<$Res>(_self.parameterSet, (value) {
    return _then(_self.copyWith(parameterSet: value));
  });
}
}


/// Adds pattern-matching-related methods to [AnnualEstimate].
extension AnnualEstimatePatterns on AnnualEstimate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualEstimate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualEstimate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualEstimate value)  $default,){
final _that = this;
switch (_that) {
case _AnnualEstimate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualEstimate value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualEstimate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(name: 'person_type')  PersonType personType,  AnnualLegend legend,  List<Disclaimer> disclaimers, @JsonKey(name: 'year_status')  AnnualYearStatus yearStatus, @JsonKey(includeIfNull: true, name: 'data_through')  Period? dataThrough, @JsonKey(includeIfNull: true)  String? deadline, @JsonKey(name: 'headline_sections')  List<AnnualComputedSectionKind> headlineSections,  List<AnnualSection> sections, @JsonKey(name: 'informative_income_over_500k')  bool informativeIncomeOver500k, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(name: 'engine_version')  EngineVersion engineVersion)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualEstimate() when $default != null:
return $default(_that.taxpayerId,_that.fiscalYear,_that.personType,_that.legend,_that.disclaimers,_that.yearStatus,_that.dataThrough,_that.deadline,_that.headlineSections,_that.sections,_that.informativeIncomeOver500k,_that.parameterSet,_that.engineVersion);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(name: 'person_type')  PersonType personType,  AnnualLegend legend,  List<Disclaimer> disclaimers, @JsonKey(name: 'year_status')  AnnualYearStatus yearStatus, @JsonKey(includeIfNull: true, name: 'data_through')  Period? dataThrough, @JsonKey(includeIfNull: true)  String? deadline, @JsonKey(name: 'headline_sections')  List<AnnualComputedSectionKind> headlineSections,  List<AnnualSection> sections, @JsonKey(name: 'informative_income_over_500k')  bool informativeIncomeOver500k, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(name: 'engine_version')  EngineVersion engineVersion)  $default,) {final _that = this;
switch (_that) {
case _AnnualEstimate():
return $default(_that.taxpayerId,_that.fiscalYear,_that.personType,_that.legend,_that.disclaimers,_that.yearStatus,_that.dataThrough,_that.deadline,_that.headlineSections,_that.sections,_that.informativeIncomeOver500k,_that.parameterSet,_that.engineVersion);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear, @JsonKey(name: 'person_type')  PersonType personType,  AnnualLegend legend,  List<Disclaimer> disclaimers, @JsonKey(name: 'year_status')  AnnualYearStatus yearStatus, @JsonKey(includeIfNull: true, name: 'data_through')  Period? dataThrough, @JsonKey(includeIfNull: true)  String? deadline, @JsonKey(name: 'headline_sections')  List<AnnualComputedSectionKind> headlineSections,  List<AnnualSection> sections, @JsonKey(name: 'informative_income_over_500k')  bool informativeIncomeOver500k, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(name: 'engine_version')  EngineVersion engineVersion)?  $default,) {final _that = this;
switch (_that) {
case _AnnualEstimate() when $default != null:
return $default(_that.taxpayerId,_that.fiscalYear,_that.personType,_that.legend,_that.disclaimers,_that.yearStatus,_that.dataThrough,_that.deadline,_that.headlineSections,_that.sections,_that.informativeIncomeOver500k,_that.parameterSet,_that.engineVersion);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualEstimate implements AnnualEstimate {
  const _AnnualEstimate({@JsonKey(name: 'taxpayer_id') required this.taxpayerId, @JsonKey(name: 'fiscal_year') required this.fiscalYear, @JsonKey(name: 'person_type') required this.personType, required this.legend, required  List<Disclaimer> disclaimers, @JsonKey(name: 'year_status') required this.yearStatus, @JsonKey(includeIfNull: true, name: 'data_through') required this.dataThrough, @JsonKey(includeIfNull: true) required this.deadline, @JsonKey(name: 'headline_sections') required  List<AnnualComputedSectionKind> headlineSections, required  List<AnnualSection> sections, @JsonKey(name: 'informative_income_over_500k') required this.informativeIncomeOver500k, @JsonKey(name: 'parameter_set') required this.parameterSet, @JsonKey(name: 'engine_version') required this.engineVersion}): _disclaimers = disclaimers,_headlineSections = headlineSections,_sections = sections;
  factory _AnnualEstimate.fromJson(Map<String, dynamic> json) => _$AnnualEstimateFromJson(json);

@override@JsonKey(name: 'taxpayer_id') final  Uuid taxpayerId;
@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
@override@JsonKey(name: 'person_type') final  PersonType personType;
@override final  AnnualLegend legend;
 final  List<Disclaimer> _disclaimers;
@override List<Disclaimer> get disclaimers {
  if (_disclaimers is EqualUnmodifiableListView) return _disclaimers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_disclaimers);
}

@override@JsonKey(name: 'year_status') final  AnnualYearStatus yearStatus;
@override@JsonKey(includeIfNull: true, name: 'data_through') final  Period? dataThrough;
@override@JsonKey(includeIfNull: true) final  String? deadline;
/// The sections whose `tax.balance` is the headline, each shown on its own (never netted).
 final  List<AnnualComputedSectionKind> _headlineSections;
/// The sections whose `tax.balance` is the headline, each shown on its own (never netted).
@override@JsonKey(name: 'headline_sections') List<AnnualComputedSectionKind> get headlineSections {
  if (_headlineSections is EqualUnmodifiableListView) return _headlineSections;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_headlineSections);
}

 final  List<AnnualSection> _sections;
@override List<AnnualSection> get sections {
  if (_sections is EqualUnmodifiableListView) return _sections;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_sections);
}

@override@JsonKey(name: 'informative_income_over_500k') final  bool informativeIncomeOver500k;
@override@JsonKey(name: 'parameter_set') final  ParameterSetRef parameterSet;
@override@JsonKey(name: 'engine_version') final  EngineVersion engineVersion;

/// Create a copy of AnnualEstimate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualEstimateCopyWith<_AnnualEstimate> get copyWith => __$AnnualEstimateCopyWithImpl<_AnnualEstimate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualEstimateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualEstimate&&(identical(other.taxpayerId, taxpayerId) || other.taxpayerId == taxpayerId)&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&(identical(other.personType, personType) || other.personType == personType)&&(identical(other.legend, legend) || other.legend == legend)&&const DeepCollectionEquality().equals(other.disclaimers, _disclaimers)&&(identical(other.yearStatus, yearStatus) || other.yearStatus == yearStatus)&&(identical(other.dataThrough, dataThrough) || other.dataThrough == dataThrough)&&(identical(other.deadline, deadline) || other.deadline == deadline)&&const DeepCollectionEquality().equals(other.headlineSections, _headlineSections)&&const DeepCollectionEquality().equals(other.sections, _sections)&&(identical(other.informativeIncomeOver500k, informativeIncomeOver500k) || other.informativeIncomeOver500k == informativeIncomeOver500k)&&(identical(other.parameterSet, parameterSet) || other.parameterSet == parameterSet)&&(identical(other.engineVersion, engineVersion) || other.engineVersion == engineVersion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,taxpayerId,fiscalYear,personType,legend,const DeepCollectionEquality().hash(_disclaimers),yearStatus,dataThrough,deadline,const DeepCollectionEquality().hash(_headlineSections),const DeepCollectionEquality().hash(_sections),informativeIncomeOver500k,parameterSet,engineVersion);
}

@override
String toString() {
    return 'AnnualEstimate(taxpayerId: $taxpayerId, fiscalYear: $fiscalYear, personType: $personType, legend: $legend, disclaimers: $disclaimers, yearStatus: $yearStatus, dataThrough: $dataThrough, deadline: $deadline, headlineSections: $headlineSections, sections: $sections, informativeIncomeOver500k: $informativeIncomeOver500k, parameterSet: $parameterSet, engineVersion: $engineVersion)';
}


}

/// @nodoc
abstract mixin class _$AnnualEstimateCopyWith<$Res> implements $AnnualEstimateCopyWith<$Res> {
  factory _$AnnualEstimateCopyWith(_AnnualEstimate value, $Res Function(_AnnualEstimate) _then) = __$AnnualEstimateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear,@JsonKey(name: 'person_type') PersonType personType, AnnualLegend legend, List<Disclaimer> disclaimers,@JsonKey(name: 'year_status') AnnualYearStatus yearStatus,@JsonKey(includeIfNull: true, name: 'data_through') Period? dataThrough,@JsonKey(includeIfNull: true) String? deadline,@JsonKey(name: 'headline_sections') List<AnnualComputedSectionKind> headlineSections, List<AnnualSection> sections,@JsonKey(name: 'informative_income_over_500k') bool informativeIncomeOver500k,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet,@JsonKey(name: 'engine_version') EngineVersion engineVersion
});


@override $ParameterSetRefCopyWith<$Res> get parameterSet;

}
/// @nodoc
class __$AnnualEstimateCopyWithImpl<$Res>
    implements _$AnnualEstimateCopyWith<$Res> {
  __$AnnualEstimateCopyWithImpl(this._self, this._then);

  final _AnnualEstimate _self;
  final $Res Function(_AnnualEstimate) _then;

/// Create a copy of AnnualEstimate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? taxpayerId = null,Object? fiscalYear = null,Object? personType = null,Object? legend = null,Object? disclaimers = null,Object? yearStatus = null,Object? dataThrough = freezed,Object? deadline = freezed,Object? headlineSections = null,Object? sections = null,Object? informativeIncomeOver500k = null,Object? parameterSet = null,Object? engineVersion = null,}) {
  return _then(_AnnualEstimate(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,personType: null == personType ? _self.personType : personType // ignore: cast_nullable_to_non_nullable
as PersonType,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as AnnualLegend,disclaimers: null == disclaimers ? _self._disclaimers : disclaimers // ignore: cast_nullable_to_non_nullable
as List<Disclaimer>,yearStatus: null == yearStatus ? _self.yearStatus : yearStatus // ignore: cast_nullable_to_non_nullable
as AnnualYearStatus,dataThrough: freezed == dataThrough ? _self.dataThrough : dataThrough // ignore: cast_nullable_to_non_nullable
as Period?,deadline: freezed == deadline ? _self.deadline : deadline // ignore: cast_nullable_to_non_nullable
as String?,headlineSections: null == headlineSections ? _self._headlineSections : headlineSections // ignore: cast_nullable_to_non_nullable
as List<AnnualComputedSectionKind>,sections: null == sections ? _self._sections : sections // ignore: cast_nullable_to_non_nullable
as List<AnnualSection>,informativeIncomeOver500k: null == informativeIncomeOver500k ? _self.informativeIncomeOver500k : informativeIncomeOver500k // ignore: cast_nullable_to_non_nullable
as bool,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as EngineVersion,
  ));
}

/// Create a copy of AnnualEstimate
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
