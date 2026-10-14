// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'history_month.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$HistoryMonth {

 Period get period; TaxpayerEstimateStatus get status;@JsonKey(name: 'has_records') bool get hasRecords;@JsonKey(includeIfNull: true) HistoryTax? get tax;@JsonKey(name: 'due_date') String get dueDate;/// Any regime of the month is pending CPA review; the app shows the strong, non-dismissible banner.
@JsonKey(name: 'cpa_review_pending') bool get cpaReviewPending;@JsonKey(name: 'is_current') bool get isCurrent;@JsonKey(name: 'records_fingerprint') RecordsFingerprint get recordsFingerprint;@JsonKey(name: 'figures_fingerprint') FiguresFingerprint get figuresFingerprint;@JsonKey(name: 'engine_version') EngineVersion get engineVersion;@JsonKey(name: 'parameter_set') ParameterSetRef get parameterSet;@JsonKey(includeIfNull: true, name: 'viewed_at') DateTime? get viewedAt;@JsonKey(name: 'changed_since_viewed') bool get changedSinceViewed;@JsonKey(includeIfNull: true, name: 'change_reason') HistoryChangeReason? get changeReason;
/// Create a copy of HistoryMonth
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$HistoryMonthCopyWith<HistoryMonth> get copyWith => _$HistoryMonthCopyWithImpl<HistoryMonth>(this as HistoryMonth, _$identity);

  /// Serializes this HistoryMonth to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as HistoryMonth;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is HistoryMonth&&(identical(other.period, _this.period) || other.period == _this.period)&&(identical(other.status, _this.status) || other.status == _this.status)&&(identical(other.hasRecords, _this.hasRecords) || other.hasRecords == _this.hasRecords)&&(identical(other.tax, _this.tax) || other.tax == _this.tax)&&(identical(other.dueDate, _this.dueDate) || other.dueDate == _this.dueDate)&&(identical(other.cpaReviewPending, _this.cpaReviewPending) || other.cpaReviewPending == _this.cpaReviewPending)&&(identical(other.isCurrent, _this.isCurrent) || other.isCurrent == _this.isCurrent)&&(identical(other.recordsFingerprint, _this.recordsFingerprint) || other.recordsFingerprint == _this.recordsFingerprint)&&(identical(other.figuresFingerprint, _this.figuresFingerprint) || other.figuresFingerprint == _this.figuresFingerprint)&&(identical(other.engineVersion, _this.engineVersion) || other.engineVersion == _this.engineVersion)&&(identical(other.parameterSet, _this.parameterSet) || other.parameterSet == _this.parameterSet)&&(identical(other.viewedAt, _this.viewedAt) || other.viewedAt == _this.viewedAt)&&(identical(other.changedSinceViewed, _this.changedSinceViewed) || other.changedSinceViewed == _this.changedSinceViewed)&&(identical(other.changeReason, _this.changeReason) || other.changeReason == _this.changeReason));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as HistoryMonth;
  return Object.hash(runtimeType,_this.period,_this.status,_this.hasRecords,_this.tax,_this.dueDate,_this.cpaReviewPending,_this.isCurrent,_this.recordsFingerprint,_this.figuresFingerprint,_this.engineVersion,_this.parameterSet,_this.viewedAt,_this.changedSinceViewed,_this.changeReason);
}

@override
String toString() {
  final _this = this as HistoryMonth;
  return 'HistoryMonth(period: ${_this.period}, status: ${_this.status}, hasRecords: ${_this.hasRecords}, tax: ${_this.tax}, dueDate: ${_this.dueDate}, cpaReviewPending: ${_this.cpaReviewPending}, isCurrent: ${_this.isCurrent}, recordsFingerprint: ${_this.recordsFingerprint}, figuresFingerprint: ${_this.figuresFingerprint}, engineVersion: ${_this.engineVersion}, parameterSet: ${_this.parameterSet}, viewedAt: ${_this.viewedAt}, changedSinceViewed: ${_this.changedSinceViewed}, changeReason: ${_this.changeReason})';
}


}

/// @nodoc
abstract mixin class $HistoryMonthCopyWith<$Res>  {
  factory $HistoryMonthCopyWith(HistoryMonth value, $Res Function(HistoryMonth) _then) = _$HistoryMonthCopyWithImpl;
@useResult
$Res call({
 Period period, TaxpayerEstimateStatus status,@JsonKey(name: 'has_records') bool hasRecords,@JsonKey(includeIfNull: true) HistoryTax? tax,@JsonKey(name: 'due_date') String dueDate,@JsonKey(name: 'cpa_review_pending') bool cpaReviewPending,@JsonKey(name: 'is_current') bool isCurrent,@JsonKey(name: 'records_fingerprint') RecordsFingerprint recordsFingerprint,@JsonKey(name: 'figures_fingerprint') FiguresFingerprint figuresFingerprint,@JsonKey(name: 'engine_version') EngineVersion engineVersion,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet,@JsonKey(includeIfNull: true, name: 'viewed_at') DateTime? viewedAt,@JsonKey(name: 'changed_since_viewed') bool changedSinceViewed,@JsonKey(includeIfNull: true, name: 'change_reason') HistoryChangeReason? changeReason
});


$HistoryTaxCopyWith<$Res>? get tax;$ParameterSetRefCopyWith<$Res> get parameterSet;

}
/// @nodoc
class _$HistoryMonthCopyWithImpl<$Res>
    implements $HistoryMonthCopyWith<$Res> {
  _$HistoryMonthCopyWithImpl(this._self, this._then);

  final HistoryMonth _self;
  final $Res Function(HistoryMonth) _then;

/// Create a copy of HistoryMonth
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? period = null,Object? status = null,Object? hasRecords = null,Object? tax = freezed,Object? dueDate = null,Object? cpaReviewPending = null,Object? isCurrent = null,Object? recordsFingerprint = null,Object? figuresFingerprint = null,Object? engineVersion = null,Object? parameterSet = null,Object? viewedAt = freezed,Object? changedSinceViewed = null,Object? changeReason = freezed,}) {
  return _then(HistoryMonth(
period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as TaxpayerEstimateStatus,hasRecords: null == hasRecords ? _self.hasRecords : hasRecords // ignore: cast_nullable_to_non_nullable
as bool,tax: freezed == tax ? _self.tax : tax // ignore: cast_nullable_to_non_nullable
as HistoryTax?,dueDate: null == dueDate ? _self.dueDate : dueDate // ignore: cast_nullable_to_non_nullable
as String,cpaReviewPending: null == cpaReviewPending ? _self.cpaReviewPending : cpaReviewPending // ignore: cast_nullable_to_non_nullable
as bool,isCurrent: null == isCurrent ? _self.isCurrent : isCurrent // ignore: cast_nullable_to_non_nullable
as bool,recordsFingerprint: null == recordsFingerprint ? _self.recordsFingerprint : recordsFingerprint // ignore: cast_nullable_to_non_nullable
as RecordsFingerprint,figuresFingerprint: null == figuresFingerprint ? _self.figuresFingerprint : figuresFingerprint // ignore: cast_nullable_to_non_nullable
as FiguresFingerprint,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as EngineVersion,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,viewedAt: freezed == viewedAt ? _self.viewedAt : viewedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,changedSinceViewed: null == changedSinceViewed ? _self.changedSinceViewed : changedSinceViewed // ignore: cast_nullable_to_non_nullable
as bool,changeReason: freezed == changeReason ? _self.changeReason : changeReason // ignore: cast_nullable_to_non_nullable
as HistoryChangeReason?,
  ));
}
/// Create a copy of HistoryMonth
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$HistoryTaxCopyWith<$Res>? get tax {
    if (_self.tax == null) {
    return null;
  }

  return $HistoryTaxCopyWith<$Res>(_self.tax!, (value) {
    return _then(_self.copyWith(tax: value));
  });
}/// Create a copy of HistoryMonth
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$ParameterSetRefCopyWith<$Res> get parameterSet {
  
  return $ParameterSetRefCopyWith<$Res>(_self.parameterSet, (value) {
    return _then(_self.copyWith(parameterSet: value));
  });
}
}


/// Adds pattern-matching-related methods to [HistoryMonth].
extension HistoryMonthPatterns on HistoryMonth {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _HistoryMonth value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _HistoryMonth() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _HistoryMonth value)  $default,){
final _that = this;
switch (_that) {
case _HistoryMonth():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _HistoryMonth value)?  $default,){
final _that = this;
switch (_that) {
case _HistoryMonth() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Period period,  TaxpayerEstimateStatus status, @JsonKey(name: 'has_records')  bool hasRecords, @JsonKey(includeIfNull: true)  HistoryTax? tax, @JsonKey(name: 'due_date')  String dueDate, @JsonKey(name: 'cpa_review_pending')  bool cpaReviewPending, @JsonKey(name: 'is_current')  bool isCurrent, @JsonKey(name: 'records_fingerprint')  RecordsFingerprint recordsFingerprint, @JsonKey(name: 'figures_fingerprint')  FiguresFingerprint figuresFingerprint, @JsonKey(name: 'engine_version')  EngineVersion engineVersion, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(includeIfNull: true, name: 'viewed_at')  DateTime? viewedAt, @JsonKey(name: 'changed_since_viewed')  bool changedSinceViewed, @JsonKey(includeIfNull: true, name: 'change_reason')  HistoryChangeReason? changeReason)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _HistoryMonth() when $default != null:
return $default(_that.period,_that.status,_that.hasRecords,_that.tax,_that.dueDate,_that.cpaReviewPending,_that.isCurrent,_that.recordsFingerprint,_that.figuresFingerprint,_that.engineVersion,_that.parameterSet,_that.viewedAt,_that.changedSinceViewed,_that.changeReason);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Period period,  TaxpayerEstimateStatus status, @JsonKey(name: 'has_records')  bool hasRecords, @JsonKey(includeIfNull: true)  HistoryTax? tax, @JsonKey(name: 'due_date')  String dueDate, @JsonKey(name: 'cpa_review_pending')  bool cpaReviewPending, @JsonKey(name: 'is_current')  bool isCurrent, @JsonKey(name: 'records_fingerprint')  RecordsFingerprint recordsFingerprint, @JsonKey(name: 'figures_fingerprint')  FiguresFingerprint figuresFingerprint, @JsonKey(name: 'engine_version')  EngineVersion engineVersion, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(includeIfNull: true, name: 'viewed_at')  DateTime? viewedAt, @JsonKey(name: 'changed_since_viewed')  bool changedSinceViewed, @JsonKey(includeIfNull: true, name: 'change_reason')  HistoryChangeReason? changeReason)  $default,) {final _that = this;
switch (_that) {
case _HistoryMonth():
return $default(_that.period,_that.status,_that.hasRecords,_that.tax,_that.dueDate,_that.cpaReviewPending,_that.isCurrent,_that.recordsFingerprint,_that.figuresFingerprint,_that.engineVersion,_that.parameterSet,_that.viewedAt,_that.changedSinceViewed,_that.changeReason);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Period period,  TaxpayerEstimateStatus status, @JsonKey(name: 'has_records')  bool hasRecords, @JsonKey(includeIfNull: true)  HistoryTax? tax, @JsonKey(name: 'due_date')  String dueDate, @JsonKey(name: 'cpa_review_pending')  bool cpaReviewPending, @JsonKey(name: 'is_current')  bool isCurrent, @JsonKey(name: 'records_fingerprint')  RecordsFingerprint recordsFingerprint, @JsonKey(name: 'figures_fingerprint')  FiguresFingerprint figuresFingerprint, @JsonKey(name: 'engine_version')  EngineVersion engineVersion, @JsonKey(name: 'parameter_set')  ParameterSetRef parameterSet, @JsonKey(includeIfNull: true, name: 'viewed_at')  DateTime? viewedAt, @JsonKey(name: 'changed_since_viewed')  bool changedSinceViewed, @JsonKey(includeIfNull: true, name: 'change_reason')  HistoryChangeReason? changeReason)?  $default,) {final _that = this;
switch (_that) {
case _HistoryMonth() when $default != null:
return $default(_that.period,_that.status,_that.hasRecords,_that.tax,_that.dueDate,_that.cpaReviewPending,_that.isCurrent,_that.recordsFingerprint,_that.figuresFingerprint,_that.engineVersion,_that.parameterSet,_that.viewedAt,_that.changedSinceViewed,_that.changeReason);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _HistoryMonth implements HistoryMonth {
  const _HistoryMonth({required this.period, required this.status, @JsonKey(name: 'has_records') required this.hasRecords, @JsonKey(includeIfNull: true) required this.tax, @JsonKey(name: 'due_date') required this.dueDate, @JsonKey(name: 'cpa_review_pending') required this.cpaReviewPending, @JsonKey(name: 'is_current') required this.isCurrent, @JsonKey(name: 'records_fingerprint') required this.recordsFingerprint, @JsonKey(name: 'figures_fingerprint') required this.figuresFingerprint, @JsonKey(name: 'engine_version') required this.engineVersion, @JsonKey(name: 'parameter_set') required this.parameterSet, @JsonKey(includeIfNull: true, name: 'viewed_at') required this.viewedAt, @JsonKey(name: 'changed_since_viewed') required this.changedSinceViewed, @JsonKey(includeIfNull: true, name: 'change_reason') required this.changeReason});
  factory _HistoryMonth.fromJson(Map<String, dynamic> json) => _$HistoryMonthFromJson(json);

@override final  Period period;
@override final  TaxpayerEstimateStatus status;
@override@JsonKey(name: 'has_records') final  bool hasRecords;
@override@JsonKey(includeIfNull: true) final  HistoryTax? tax;
@override@JsonKey(name: 'due_date') final  String dueDate;
/// Any regime of the month is pending CPA review; the app shows the strong, non-dismissible banner.
@override@JsonKey(name: 'cpa_review_pending') final  bool cpaReviewPending;
@override@JsonKey(name: 'is_current') final  bool isCurrent;
@override@JsonKey(name: 'records_fingerprint') final  RecordsFingerprint recordsFingerprint;
@override@JsonKey(name: 'figures_fingerprint') final  FiguresFingerprint figuresFingerprint;
@override@JsonKey(name: 'engine_version') final  EngineVersion engineVersion;
@override@JsonKey(name: 'parameter_set') final  ParameterSetRef parameterSet;
@override@JsonKey(includeIfNull: true, name: 'viewed_at') final  DateTime? viewedAt;
@override@JsonKey(name: 'changed_since_viewed') final  bool changedSinceViewed;
@override@JsonKey(includeIfNull: true, name: 'change_reason') final  HistoryChangeReason? changeReason;

/// Create a copy of HistoryMonth
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$HistoryMonthCopyWith<_HistoryMonth> get copyWith => __$HistoryMonthCopyWithImpl<_HistoryMonth>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$HistoryMonthToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _HistoryMonth&&(identical(other.period, period) || other.period == period)&&(identical(other.status, status) || other.status == status)&&(identical(other.hasRecords, hasRecords) || other.hasRecords == hasRecords)&&(identical(other.tax, tax) || other.tax == tax)&&(identical(other.dueDate, dueDate) || other.dueDate == dueDate)&&(identical(other.cpaReviewPending, cpaReviewPending) || other.cpaReviewPending == cpaReviewPending)&&(identical(other.isCurrent, isCurrent) || other.isCurrent == isCurrent)&&(identical(other.recordsFingerprint, recordsFingerprint) || other.recordsFingerprint == recordsFingerprint)&&(identical(other.figuresFingerprint, figuresFingerprint) || other.figuresFingerprint == figuresFingerprint)&&(identical(other.engineVersion, engineVersion) || other.engineVersion == engineVersion)&&(identical(other.parameterSet, parameterSet) || other.parameterSet == parameterSet)&&(identical(other.viewedAt, viewedAt) || other.viewedAt == viewedAt)&&(identical(other.changedSinceViewed, changedSinceViewed) || other.changedSinceViewed == changedSinceViewed)&&(identical(other.changeReason, changeReason) || other.changeReason == changeReason));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,period,status,hasRecords,tax,dueDate,cpaReviewPending,isCurrent,recordsFingerprint,figuresFingerprint,engineVersion,parameterSet,viewedAt,changedSinceViewed,changeReason);
}

@override
String toString() {
    return 'HistoryMonth(period: $period, status: $status, hasRecords: $hasRecords, tax: $tax, dueDate: $dueDate, cpaReviewPending: $cpaReviewPending, isCurrent: $isCurrent, recordsFingerprint: $recordsFingerprint, figuresFingerprint: $figuresFingerprint, engineVersion: $engineVersion, parameterSet: $parameterSet, viewedAt: $viewedAt, changedSinceViewed: $changedSinceViewed, changeReason: $changeReason)';
}


}

/// @nodoc
abstract mixin class _$HistoryMonthCopyWith<$Res> implements $HistoryMonthCopyWith<$Res> {
  factory _$HistoryMonthCopyWith(_HistoryMonth value, $Res Function(_HistoryMonth) _then) = __$HistoryMonthCopyWithImpl;
@override @useResult
$Res call({
 Period period, TaxpayerEstimateStatus status,@JsonKey(name: 'has_records') bool hasRecords,@JsonKey(includeIfNull: true) HistoryTax? tax,@JsonKey(name: 'due_date') String dueDate,@JsonKey(name: 'cpa_review_pending') bool cpaReviewPending,@JsonKey(name: 'is_current') bool isCurrent,@JsonKey(name: 'records_fingerprint') RecordsFingerprint recordsFingerprint,@JsonKey(name: 'figures_fingerprint') FiguresFingerprint figuresFingerprint,@JsonKey(name: 'engine_version') EngineVersion engineVersion,@JsonKey(name: 'parameter_set') ParameterSetRef parameterSet,@JsonKey(includeIfNull: true, name: 'viewed_at') DateTime? viewedAt,@JsonKey(name: 'changed_since_viewed') bool changedSinceViewed,@JsonKey(includeIfNull: true, name: 'change_reason') HistoryChangeReason? changeReason
});


@override $HistoryTaxCopyWith<$Res>? get tax;@override $ParameterSetRefCopyWith<$Res> get parameterSet;

}
/// @nodoc
class __$HistoryMonthCopyWithImpl<$Res>
    implements _$HistoryMonthCopyWith<$Res> {
  __$HistoryMonthCopyWithImpl(this._self, this._then);

  final _HistoryMonth _self;
  final $Res Function(_HistoryMonth) _then;

/// Create a copy of HistoryMonth
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? period = null,Object? status = null,Object? hasRecords = null,Object? tax = freezed,Object? dueDate = null,Object? cpaReviewPending = null,Object? isCurrent = null,Object? recordsFingerprint = null,Object? figuresFingerprint = null,Object? engineVersion = null,Object? parameterSet = null,Object? viewedAt = freezed,Object? changedSinceViewed = null,Object? changeReason = freezed,}) {
  return _then(_HistoryMonth(
period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as TaxpayerEstimateStatus,hasRecords: null == hasRecords ? _self.hasRecords : hasRecords // ignore: cast_nullable_to_non_nullable
as bool,tax: freezed == tax ? _self.tax : tax // ignore: cast_nullable_to_non_nullable
as HistoryTax?,dueDate: null == dueDate ? _self.dueDate : dueDate // ignore: cast_nullable_to_non_nullable
as String,cpaReviewPending: null == cpaReviewPending ? _self.cpaReviewPending : cpaReviewPending // ignore: cast_nullable_to_non_nullable
as bool,isCurrent: null == isCurrent ? _self.isCurrent : isCurrent // ignore: cast_nullable_to_non_nullable
as bool,recordsFingerprint: null == recordsFingerprint ? _self.recordsFingerprint : recordsFingerprint // ignore: cast_nullable_to_non_nullable
as RecordsFingerprint,figuresFingerprint: null == figuresFingerprint ? _self.figuresFingerprint : figuresFingerprint // ignore: cast_nullable_to_non_nullable
as FiguresFingerprint,engineVersion: null == engineVersion ? _self.engineVersion : engineVersion // ignore: cast_nullable_to_non_nullable
as EngineVersion,parameterSet: null == parameterSet ? _self.parameterSet : parameterSet // ignore: cast_nullable_to_non_nullable
as ParameterSetRef,viewedAt: freezed == viewedAt ? _self.viewedAt : viewedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,changedSinceViewed: null == changedSinceViewed ? _self.changedSinceViewed : changedSinceViewed // ignore: cast_nullable_to_non_nullable
as bool,changeReason: freezed == changeReason ? _self.changeReason : changeReason // ignore: cast_nullable_to_non_nullable
as HistoryChangeReason?,
  ));
}

/// Create a copy of HistoryMonth
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$HistoryTaxCopyWith<$Res>? get tax {
    if (_self.tax == null) {
    return null;
  }

  return $HistoryTaxCopyWith<$Res>(_self.tax!, (value) {
    return _then(_self.copyWith(tax: value));
  });
}/// Create a copy of HistoryMonth
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
