// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'taxpayer_estimate.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxpayerEstimate {

@JsonKey(name: 'taxpayer_id') Uuid get taxpayerId; Period get period;@JsonKey(name: 'due_date') String get dueDate; Legend get legend; List<Disclaimer> get disclaimers; List<RegimeEstimate> get regimes; EstimateTotals get totals;@JsonKey(name: 'parameter_set') ParameterSetRef get parameterSet;@JsonKey(name: 'records_fingerprint') RecordsFingerprint get recordsFingerprint;@JsonKey(name: 'engine_version') String get engineVersion;
/// Create a copy of TaxpayerEstimate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxpayerEstimateCopyWith<TaxpayerEstimate> get copyWith => _$TaxpayerEstimateCopyWithImpl<TaxpayerEstimate>(this as TaxpayerEstimate, _$identity);

  /// Serializes this TaxpayerEstimate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxpayerEstimate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxpayerEstimate&&(identical(other.taxpayerId, _this.taxpayerId) || other.taxpayerId == _this.taxpayerId)&&(identical(other.period, _this.period) || other.period == _this.period)&&(identical(other.dueDate, _this.dueDate) || other.dueDate == _this.dueDate)&&(identical(other.legend, _this.legend) || other.legend == _this.legend)&&const DeepCollectionEquality().equals(other.disclaimers, _this.disclaimers)&&const DeepCollectionEquality().equals(other.regimes, _this.regimes)&&(identical(other.totals, _this.totals) || other.totals == _this.totals)&&(identical(other.parameterSet, _this.parameterSet) || other.parameterSet == _this.parameterSet)&&(identical(other.recordsFingerprint, _this.recordsFingerprint) || other.recordsFingerprint == _this.recordsFingerprint)&&(identical(other.engineVersion, _this.engineVersion) || other.engineVersion == _this.engineVersion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxpayerEstimate;
  return Object.hash(runtimeType,_this.taxpayerId,_this.period,_this.dueDate,_this.legend,const DeepCollectionEquality().hash(_this.disclaimers),const DeepCollectionEquality().hash(_this.regimes),_this.totals,_this.parameterSet,_this.recordsFingerprint,_this.engineVersion);
}

@override
String toString() {
  final _this = this as TaxpayerEstimate;
  return 'TaxpayerEstimate(taxpayerId: ${_this.taxpayerId}, period: ${_this.period}, dueDate: ${_this.dueDate}, legend: ${_this.legend}, disclaimers: ${_this.disclaimers}, regimes: ${_this.regimes}, totals: ${_this.totals}, parameterSet: ${_this.parameterSet}, recordsFingerprint: ${_this.recordsFingerprint}, engineVersion: ${_this.engineVersion})';
}


}

/// @nodoc
abstract mixin class $TaxpayerEstimateCopyWith<$Res>  {
  factory $TaxpayerEstimateCopyWith(TaxpayerEstimate value, $Res Function(TaxpayerEstimate) _then) = _$TaxpayerEstimateCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId, Period period,@JsonKey(name: 'due_date') String dueDate, Legend legend, List<Disclaimer> disclaimers, List<RegimeEstimate> regimes, EstimateTotals totals,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet,@JsonKey(name: 'records_fingerprint') RecordsFingerprint recordsFingerprint,@JsonKey(name: 'engine_version') String engineVersion
});


$EstimateTotalsCopyWith<$Res> get totals;$ParameterSetRefCopyWith<$Res> get parameterSet;

}
/// @nodoc
class _$TaxpayerEstimateCopyWithImpl<$Res>
    implements $TaxpayerEstimateCopyWith<$Res> {
  _$TaxpayerEstimateCopyWithImpl(this._self, this._then);

  final TaxpayerEstimate _self;
  final $Res Function(TaxpayerEstimate) _then;

/// Create a copy of TaxpayerEstimate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? taxpayerId = null,Object? period = null,Object? dueDate = null,Object? legend = null,Object? disclaimers = null,Object? regimes = null,Object? totals = null,Object? parameterSet = null,Object? recordsFingerprint = null,Object? engineVersion = null,}) {
  return _then(TaxpayerEstimate(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,dueDate: null == dueDate ? _self.dueDate : dueDate // ignore: cast_nullable_to_non_nullable
as String,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,disclaimers: null == disclaimers ? _self.disclaimers : disclaimers // ignore: cast_nullable_to_non_nullable
as List<Disclaimer>,regimes: null == regimes ? _self.regimes : regimes // ignore: cast_nullable_to_non_nullable
as List<RegimeEstimate>,totals: null == totals ? _self.totals : totals // ignore: cast_nullable_to_non_nullable
as EstimateTotals,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,recordsFingerprint: null == recordsFingerprint ? _self.recordsFingerprint : recordsFingerprint // ignore: cast_nullable_to_non_nullable
as RecordsFingerprint,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as String,
  ));
}
/// Create a copy of TaxpayerEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$EstimateTotalsCopyWith<$Res> get totals {
  
  return $EstimateTotalsCopyWith<$Res>(_self.totals, (value) {
    return _then(_self.copyWith(totals: value));
  });
}/// Create a copy of TaxpayerEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ParameterSetRefCopyWith<$Res> get parameterSet {
  
  return $ParameterSetRefCopyWith<$Res>(_self.parameterSet, (value) {
    return _then(_self.copyWith(parameterSet: value));
  });
}
}


/// Adds pattern-matching-related methods to [TaxpayerEstimate].
extension TaxpayerEstimatePatterns on TaxpayerEstimate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxpayerEstimate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxpayerEstimate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxpayerEstimate value)  $default,){
final _that = this;
switch (_that) {
case _TaxpayerEstimate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxpayerEstimate value)?  $default,){
final _that = this;
switch (_that) {
case _TaxpayerEstimate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  Period period, @JsonKey(name: 'due_date')  String dueDate,  Legend legend,  List<Disclaimer> disclaimers,  List<RegimeEstimate> regimes,  EstimateTotals totals, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(name: 'records_fingerprint')  RecordsFingerprint recordsFingerprint, @JsonKey(name: 'engine_version')  String engineVersion)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxpayerEstimate() when $default != null:
return $default(_that.taxpayerId,_that.period,_that.dueDate,_that.legend,_that.disclaimers,_that.regimes,_that.totals,_that.parameterSet,_that.recordsFingerprint,_that.engineVersion);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  Period period, @JsonKey(name: 'due_date')  String dueDate,  Legend legend,  List<Disclaimer> disclaimers,  List<RegimeEstimate> regimes,  EstimateTotals totals, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(name: 'records_fingerprint')  RecordsFingerprint recordsFingerprint, @JsonKey(name: 'engine_version')  String engineVersion)  $default,) {final _that = this;
switch (_that) {
case _TaxpayerEstimate():
return $default(_that.taxpayerId,_that.period,_that.dueDate,_that.legend,_that.disclaimers,_that.regimes,_that.totals,_that.parameterSet,_that.recordsFingerprint,_that.engineVersion);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  Period period, @JsonKey(name: 'due_date')  String dueDate,  Legend legend,  List<Disclaimer> disclaimers,  List<RegimeEstimate> regimes,  EstimateTotals totals, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(name: 'records_fingerprint')  RecordsFingerprint recordsFingerprint, @JsonKey(name: 'engine_version')  String engineVersion)?  $default,) {final _that = this;
switch (_that) {
case _TaxpayerEstimate() when $default != null:
return $default(_that.taxpayerId,_that.period,_that.dueDate,_that.legend,_that.disclaimers,_that.regimes,_that.totals,_that.parameterSet,_that.recordsFingerprint,_that.engineVersion);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxpayerEstimate implements TaxpayerEstimate {
  const _TaxpayerEstimate({@JsonKey(name: 'taxpayer_id') required this.taxpayerId, required this.period, @JsonKey(name: 'due_date') required this.dueDate, required this.legend, required  List<Disclaimer> disclaimers, required  List<RegimeEstimate> regimes, required this.totals, @JsonKey(name: 'parameter_set') required this.parameterSet, @JsonKey(name: 'records_fingerprint') required this.recordsFingerprint, @JsonKey(name: 'engine_version') required this.engineVersion}): _disclaimers = disclaimers,_regimes = regimes;
  factory _TaxpayerEstimate.fromJson(Map<String, dynamic> json) => _$TaxpayerEstimateFromJson(json);

@override@JsonKey(name: 'taxpayer_id') final  Uuid taxpayerId;
@override final  Period period;
@override@JsonKey(name: 'due_date') final  String dueDate;
@override final  Legend legend;
 final  List<Disclaimer> _disclaimers;
@override List<Disclaimer> get disclaimers {
  if (_disclaimers is EqualUnmodifiableListView) return _disclaimers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_disclaimers);
}

 final  List<RegimeEstimate> _regimes;
@override List<RegimeEstimate> get regimes {
  if (_regimes is EqualUnmodifiableListView) return _regimes;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_regimes);
}

@override final  EstimateTotals totals;
@override@JsonKey(name: 'parameter_set') final  ParameterSetRef parameterSet;
@override@JsonKey(name: 'records_fingerprint') final  RecordsFingerprint recordsFingerprint;
@override@JsonKey(name: 'engine_version') final  String engineVersion;

/// Create a copy of TaxpayerEstimate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxpayerEstimateCopyWith<_TaxpayerEstimate> get copyWith => __$TaxpayerEstimateCopyWithImpl<_TaxpayerEstimate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxpayerEstimateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxpayerEstimate&&(identical(other.taxpayerId, taxpayerId) || other.taxpayerId == taxpayerId)&&(identical(other.period, period) || other.period == period)&&(identical(other.dueDate, dueDate) || other.dueDate == dueDate)&&(identical(other.legend, legend) || other.legend == legend)&&const DeepCollectionEquality().equals(other.disclaimers, _disclaimers)&&const DeepCollectionEquality().equals(other.regimes, _regimes)&&(identical(other.totals, totals) || other.totals == totals)&&(identical(other.parameterSet, parameterSet) || other.parameterSet == parameterSet)&&(identical(other.recordsFingerprint, recordsFingerprint) || other.recordsFingerprint == recordsFingerprint)&&(identical(other.engineVersion, engineVersion) || other.engineVersion == engineVersion));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,taxpayerId,period,dueDate,legend,const DeepCollectionEquality().hash(_disclaimers),const DeepCollectionEquality().hash(_regimes),totals,parameterSet,recordsFingerprint,engineVersion);
}

@override
String toString() {
    return 'TaxpayerEstimate(taxpayerId: $taxpayerId, period: $period, dueDate: $dueDate, legend: $legend, disclaimers: $disclaimers, regimes: $regimes, totals: $totals, parameterSet: $parameterSet, recordsFingerprint: $recordsFingerprint, engineVersion: $engineVersion)';
}


}

/// @nodoc
abstract mixin class _$TaxpayerEstimateCopyWith<$Res> implements $TaxpayerEstimateCopyWith<$Res> {
  factory _$TaxpayerEstimateCopyWith(_TaxpayerEstimate value, $Res Function(_TaxpayerEstimate) _then) = __$TaxpayerEstimateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId, Period period,@JsonKey(name: 'due_date') String dueDate, Legend legend, List<Disclaimer> disclaimers, List<RegimeEstimate> regimes, EstimateTotals totals,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet,@JsonKey(name: 'records_fingerprint') RecordsFingerprint recordsFingerprint,@JsonKey(name: 'engine_version') String engineVersion
});


@override $EstimateTotalsCopyWith<$Res> get totals;@override $ParameterSetRefCopyWith<$Res> get parameterSet;

}
/// @nodoc
class __$TaxpayerEstimateCopyWithImpl<$Res>
    implements _$TaxpayerEstimateCopyWith<$Res> {
  __$TaxpayerEstimateCopyWithImpl(this._self, this._then);

  final _TaxpayerEstimate _self;
  final $Res Function(_TaxpayerEstimate) _then;

/// Create a copy of TaxpayerEstimate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? taxpayerId = null,Object? period = null,Object? dueDate = null,Object? legend = null,Object? disclaimers = null,Object? regimes = null,Object? totals = null,Object? parameterSet = null,Object? recordsFingerprint = null,Object? engineVersion = null,}) {
  return _then(_TaxpayerEstimate(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,dueDate: null == dueDate ? _self.dueDate : dueDate // ignore: cast_nullable_to_non_nullable
as String,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,disclaimers: null == disclaimers ? _self._disclaimers : disclaimers // ignore: cast_nullable_to_non_nullable
as List<Disclaimer>,regimes: null == regimes ? _self._regimes : regimes // ignore: cast_nullable_to_non_nullable
as List<RegimeEstimate>,totals: null == totals ? _self.totals : totals // ignore: cast_nullable_to_non_nullable
as EstimateTotals,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,recordsFingerprint: null == recordsFingerprint ? _self.recordsFingerprint : recordsFingerprint // ignore: cast_nullable_to_non_nullable
as RecordsFingerprint,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

/// Create a copy of TaxpayerEstimate
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$EstimateTotalsCopyWith<$Res> get totals {
  
  return $EstimateTotalsCopyWith<$Res>(_self.totals, (value) {
    return _then(_self.copyWith(totals: value));
  });
}/// Create a copy of TaxpayerEstimate
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
