// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'dashboard.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Dashboard {

@JsonKey(name: 'taxpayer_id') Uuid get taxpayerId; Period get period;@JsonKey(name: 'is_empty') bool get isEmpty; DashboardMonth get month;@JsonKey(includeIfNull: true) TaxpayerEstimate? get estimate; List<TrendPoint> get trend; AccessSummary get access;@JsonKey(includeIfNull: false) Legend? get legend;
/// Create a copy of Dashboard
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DashboardCopyWith<Dashboard> get copyWith => _$DashboardCopyWithImpl<Dashboard>(this as Dashboard, _$identity);

  /// Serializes this Dashboard to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Dashboard;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Dashboard&&(identical(other.taxpayerId, _this.taxpayerId) || other.taxpayerId == _this.taxpayerId)&&(identical(other.period, _this.period) || other.period == _this.period)&&(identical(other.isEmpty, _this.isEmpty) || other.isEmpty == _this.isEmpty)&&(identical(other.month, _this.month) || other.month == _this.month)&&(identical(other.estimate, _this.estimate) || other.estimate == _this.estimate)&&const DeepCollectionEquality().equals(other.trend, _this.trend)&&(identical(other.access, _this.access) || other.access == _this.access)&&(identical(other.legend, _this.legend) || other.legend == _this.legend));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Dashboard;
  return Object.hash(runtimeType,_this.taxpayerId,_this.period,_this.isEmpty,_this.month,_this.estimate,const DeepCollectionEquality().hash(_this.trend),_this.access,_this.legend);
}

@override
String toString() {
  final _this = this as Dashboard;
  return 'Dashboard(taxpayerId: ${_this.taxpayerId}, period: ${_this.period}, isEmpty: ${_this.isEmpty}, month: ${_this.month}, estimate: ${_this.estimate}, trend: ${_this.trend}, access: ${_this.access}, legend: ${_this.legend})';
}


}

/// @nodoc
abstract mixin class $DashboardCopyWith<$Res>  {
  factory $DashboardCopyWith(Dashboard value, $Res Function(Dashboard) _then) = _$DashboardCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId, Period period,@JsonKey(name: 'is_empty') bool isEmpty, DashboardMonth month,@JsonKey(includeIfNull: true) TaxpayerEstimate? estimate, List<TrendPoint> trend, AccessSummary access,@JsonKey(includeIfNull: false) Legend? legend
});


$DashboardMonthCopyWith<$Res> get month;$TaxpayerEstimateCopyWith<$Res>? get estimate;$AccessSummaryCopyWith<$Res> get access;

}
/// @nodoc
class _$DashboardCopyWithImpl<$Res>
    implements $DashboardCopyWith<$Res> {
  _$DashboardCopyWithImpl(this._self, this._then);

  final Dashboard _self;
  final $Res Function(Dashboard) _then;

/// Create a copy of Dashboard
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? taxpayerId = null,Object? period = null,Object? isEmpty = null,Object? month = null,Object? estimate = freezed,Object? trend = null,Object? access = null,Object? legend = freezed,}) {
  return _then(Dashboard(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,isEmpty: null == isEmpty ? _self.isEmpty : isEmpty // ignore: cast_nullable_to_non_nullable
as bool,month: null == month ? _self.month : month // ignore: cast_nullable_to_non_nullable
as DashboardMonth,estimate: freezed == estimate ? _self.estimate : estimate // ignore: cast_nullable_to_non_nullable
as TaxpayerEstimate?,trend: null == trend ? _self.trend : trend // ignore: cast_nullable_to_non_nullable
as List<TrendPoint>,access: null == access ? _self.access : access // ignore: cast_nullable_to_non_nullable
as AccessSummary,legend: freezed == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend?,
  ));
}
/// Create a copy of Dashboard
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$DashboardMonthCopyWith<$Res> get month {
  
  return $DashboardMonthCopyWith<$Res>(_self.month, (value) {
    return _then(_self.copyWith(month: value));
  });
}/// Create a copy of Dashboard
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TaxpayerEstimateCopyWith<$Res>? get estimate {
    if (_self.estimate == null) {
    return null;
  }

  return $TaxpayerEstimateCopyWith<$Res>(_self.estimate!, (value) {
    return _then(_self.copyWith(estimate: value));
  });
}/// Create a copy of Dashboard
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AccessSummaryCopyWith<$Res> get access {
  
  return $AccessSummaryCopyWith<$Res>(_self.access, (value) {
    return _then(_self.copyWith(access: value));
  });
}
}


/// Adds pattern-matching-related methods to [Dashboard].
extension DashboardPatterns on Dashboard {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Dashboard value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Dashboard() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Dashboard value)  $default,){
final _that = this;
switch (_that) {
case _Dashboard():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Dashboard value)?  $default,){
final _that = this;
switch (_that) {
case _Dashboard() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  Period period, @JsonKey(name: 'is_empty')  bool isEmpty,  DashboardMonth month, @JsonKey(includeIfNull: true)  TaxpayerEstimate? estimate,  List<TrendPoint> trend,  AccessSummary access, @JsonKey(includeIfNull: false)  Legend? legend)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Dashboard() when $default != null:
return $default(_that.taxpayerId,_that.period,_that.isEmpty,_that.month,_that.estimate,_that.trend,_that.access,_that.legend);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  Period period, @JsonKey(name: 'is_empty')  bool isEmpty,  DashboardMonth month, @JsonKey(includeIfNull: true)  TaxpayerEstimate? estimate,  List<TrendPoint> trend,  AccessSummary access, @JsonKey(includeIfNull: false)  Legend? legend)  $default,) {final _that = this;
switch (_that) {
case _Dashboard():
return $default(_that.taxpayerId,_that.period,_that.isEmpty,_that.month,_that.estimate,_that.trend,_that.access,_that.legend);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId,  Period period, @JsonKey(name: 'is_empty')  bool isEmpty,  DashboardMonth month, @JsonKey(includeIfNull: true)  TaxpayerEstimate? estimate,  List<TrendPoint> trend,  AccessSummary access, @JsonKey(includeIfNull: false)  Legend? legend)?  $default,) {final _that = this;
switch (_that) {
case _Dashboard() when $default != null:
return $default(_that.taxpayerId,_that.period,_that.isEmpty,_that.month,_that.estimate,_that.trend,_that.access,_that.legend);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Dashboard implements Dashboard {
  const _Dashboard({@JsonKey(name: 'taxpayer_id') required this.taxpayerId, required this.period, @JsonKey(name: 'is_empty') required this.isEmpty, required this.month, @JsonKey(includeIfNull: true) required this.estimate, required  List<TrendPoint> trend, required this.access, @JsonKey(includeIfNull: false) this.legend}): _trend = trend;
  factory _Dashboard.fromJson(Map<String, dynamic> json) => _$DashboardFromJson(json);

@override@JsonKey(name: 'taxpayer_id') final  Uuid taxpayerId;
@override final  Period period;
@override@JsonKey(name: 'is_empty') final  bool isEmpty;
@override final  DashboardMonth month;
@override@JsonKey(includeIfNull: true) final  TaxpayerEstimate? estimate;
 final  List<TrendPoint> _trend;
@override List<TrendPoint> get trend {
  if (_trend is EqualUnmodifiableListView) return _trend;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_trend);
}

@override final  AccessSummary access;
@override@JsonKey(includeIfNull: false) final  Legend? legend;

/// Create a copy of Dashboard
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DashboardCopyWith<_Dashboard> get copyWith => __$DashboardCopyWithImpl<_Dashboard>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DashboardToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Dashboard&&(identical(other.taxpayerId, taxpayerId) || other.taxpayerId == taxpayerId)&&(identical(other.period, period) || other.period == period)&&(identical(other.isEmpty, isEmpty) || other.isEmpty == isEmpty)&&(identical(other.month, month) || other.month == month)&&(identical(other.estimate, estimate) || other.estimate == estimate)&&const DeepCollectionEquality().equals(other.trend, _trend)&&(identical(other.access, access) || other.access == access)&&(identical(other.legend, legend) || other.legend == legend));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,taxpayerId,period,isEmpty,month,estimate,const DeepCollectionEquality().hash(_trend),access,legend);
}

@override
String toString() {
    return 'Dashboard(taxpayerId: $taxpayerId, period: $period, isEmpty: $isEmpty, month: $month, estimate: $estimate, trend: $trend, access: $access, legend: $legend)';
}


}

/// @nodoc
abstract mixin class _$DashboardCopyWith<$Res> implements $DashboardCopyWith<$Res> {
  factory _$DashboardCopyWith(_Dashboard value, $Res Function(_Dashboard) _then) = __$DashboardCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId, Period period,@JsonKey(name: 'is_empty') bool isEmpty, DashboardMonth month,@JsonKey(includeIfNull: true) TaxpayerEstimate? estimate, List<TrendPoint> trend, AccessSummary access,@JsonKey(includeIfNull: false) Legend? legend
});


@override $DashboardMonthCopyWith<$Res> get month;@override $TaxpayerEstimateCopyWith<$Res>? get estimate;@override $AccessSummaryCopyWith<$Res> get access;

}
/// @nodoc
class __$DashboardCopyWithImpl<$Res>
    implements _$DashboardCopyWith<$Res> {
  __$DashboardCopyWithImpl(this._self, this._then);

  final _Dashboard _self;
  final $Res Function(_Dashboard) _then;

/// Create a copy of Dashboard
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? taxpayerId = null,Object? period = null,Object? isEmpty = null,Object? month = null,Object? estimate = freezed,Object? trend = null,Object? access = null,Object? legend = freezed,}) {
  return _then(_Dashboard(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,isEmpty: null == isEmpty ? _self.isEmpty : isEmpty // ignore: cast_nullable_to_non_nullable
as bool,month: null == month ? _self.month : month // ignore: cast_nullable_to_non_nullable
as DashboardMonth,estimate: freezed == estimate ? _self.estimate : estimate // ignore: cast_nullable_to_non_nullable
as TaxpayerEstimate?,trend: null == trend ? _self._trend : trend // ignore: cast_nullable_to_non_nullable
as List<TrendPoint>,access: null == access ? _self.access : access // ignore: cast_nullable_to_non_nullable
as AccessSummary,legend: freezed == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend?,
  ));
}

/// Create a copy of Dashboard
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$DashboardMonthCopyWith<$Res> get month {
  
  return $DashboardMonthCopyWith<$Res>(_self.month, (value) {
    return _then(_self.copyWith(month: value));
  });
}/// Create a copy of Dashboard
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TaxpayerEstimateCopyWith<$Res>? get estimate {
    if (_self.estimate == null) {
    return null;
  }

  return $TaxpayerEstimateCopyWith<$Res>(_self.estimate!, (value) {
    return _then(_self.copyWith(estimate: value));
  });
}/// Create a copy of Dashboard
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AccessSummaryCopyWith<$Res> get access {
  
  return $AccessSummaryCopyWith<$Res>(_self.access, (value) {
    return _then(_self.copyWith(access: value));
  });
}
}

// dart format on
