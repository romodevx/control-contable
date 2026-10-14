// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'history_year.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$HistoryYear {

@JsonKey(name: 'taxpayer_id') Uuid get taxpayerId;@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear; List<HistoryMonth> get months;@JsonKey(name: 'year_totals') HistoryYearTotals get yearTotals; Legend get legend; List<Disclaimer> get disclaimers;
/// Create a copy of HistoryYear
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$HistoryYearCopyWith<HistoryYear> get copyWith => _$HistoryYearCopyWithImpl<HistoryYear>(this as HistoryYear, _$identity);

  /// Serializes this HistoryYear to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as HistoryYear;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is HistoryYear&&(identical(other.taxpayerId, _this.taxpayerId) || other.taxpayerId == _this.taxpayerId)&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&const DeepCollectionEquality().equals(other.months, _this.months)&&(identical(other.yearTotals, _this.yearTotals) || other.yearTotals == _this.yearTotals)&&(identical(other.legend, _this.legend) || other.legend == _this.legend)&&const DeepCollectionEquality().equals(other.disclaimers, _this.disclaimers));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as HistoryYear;
  return Object.hash(runtimeType,_this.taxpayerId,_this.fiscalYear,const DeepCollectionEquality().hash(_this.months),_this.yearTotals,_this.legend,const DeepCollectionEquality().hash(_this.disclaimers));
}

@override
String toString() {
  final _this = this as HistoryYear;
  return 'HistoryYear(taxpayerId: ${_this.taxpayerId}, fiscalYear: ${_this.fiscalYear}, months: ${_this.months}, yearTotals: ${_this.yearTotals}, legend: ${_this.legend}, disclaimers: ${_this.disclaimers})';
}


}

/// @nodoc
abstract mixin class $HistoryYearCopyWith<$Res>  {
  factory $HistoryYearCopyWith(HistoryYear value, $Res Function(HistoryYear) _then) = _$HistoryYearCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, List<HistoryMonth> months,@JsonKey(name: 'year_totals') HistoryYearTotals yearTotals, Legend legend, List<Disclaimer> disclaimers
});


$HistoryYearTotalsCopyWith<$Res> get yearTotals;

}
/// @nodoc
class _$HistoryYearCopyWithImpl<$Res>
    implements $HistoryYearCopyWith<$Res> {
  _$HistoryYearCopyWithImpl(this._self, this._then);

  final HistoryYear _self;
  final $Res Function(HistoryYear) _then;

/// Create a copy of HistoryYear
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? taxpayerId = null,Object? fiscalYear = null,Object? months = null,Object? yearTotals = null,Object? legend = null,Object? disclaimers = null,}) {
  return _then(HistoryYear(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,months: null == months ? _self.months : months // ignore: cast_nullable_to_non_nullable
as List<HistoryMonth>,yearTotals: null == yearTotals ? _self.yearTotals : yearTotals // ignore: cast_nullable_to_non_nullable
as HistoryYearTotals,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,disclaimers: null == disclaimers ? _self.disclaimers : disclaimers // ignore: cast_nullable_to_non_nullable
as List<Disclaimer>,
  ));
}
/// Create a copy of HistoryYear
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$HistoryYearTotalsCopyWith<$Res> get yearTotals {
  
  return $HistoryYearTotalsCopyWith<$Res>(_self.yearTotals, (value) {
    return _then(_self.copyWith(yearTotals: value));
  });
}
}


/// Adds pattern-matching-related methods to [HistoryYear].
extension HistoryYearPatterns on HistoryYear {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _HistoryYear value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _HistoryYear() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _HistoryYear value)  $default,){
final _that = this;
switch (_that) {
case _HistoryYear():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _HistoryYear value)?  $default,){
final _that = this;
switch (_that) {
case _HistoryYear() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<HistoryMonth> months, @JsonKey(name: 'year_totals')  HistoryYearTotals yearTotals,  Legend legend,  List<Disclaimer> disclaimers)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _HistoryYear() when $default != null:
return $default(_that.taxpayerId,_that.fiscalYear,_that.months,_that.yearTotals,_that.legend,_that.disclaimers);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<HistoryMonth> months, @JsonKey(name: 'year_totals')  HistoryYearTotals yearTotals,  Legend legend,  List<Disclaimer> disclaimers)  $default,) {final _that = this;
switch (_that) {
case _HistoryYear():
return $default(_that.taxpayerId,_that.fiscalYear,_that.months,_that.yearTotals,_that.legend,_that.disclaimers);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<HistoryMonth> months, @JsonKey(name: 'year_totals')  HistoryYearTotals yearTotals,  Legend legend,  List<Disclaimer> disclaimers)?  $default,) {final _that = this;
switch (_that) {
case _HistoryYear() when $default != null:
return $default(_that.taxpayerId,_that.fiscalYear,_that.months,_that.yearTotals,_that.legend,_that.disclaimers);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _HistoryYear implements HistoryYear {
  const _HistoryYear({@JsonKey(name: 'taxpayer_id') required this.taxpayerId, @JsonKey(name: 'fiscal_year') required this.fiscalYear, required  List<HistoryMonth> months, @JsonKey(name: 'year_totals') required this.yearTotals, required this.legend, required  List<Disclaimer> disclaimers}): _months = months,_disclaimers = disclaimers;
  factory _HistoryYear.fromJson(Map<String, dynamic> json) => _$HistoryYearFromJson(json);

@override@JsonKey(name: 'taxpayer_id') final  Uuid taxpayerId;
@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
 final  List<HistoryMonth> _months;
@override List<HistoryMonth> get months {
  if (_months is EqualUnmodifiableListView) return _months;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_months);
}

@override@JsonKey(name: 'year_totals') final  HistoryYearTotals yearTotals;
@override final  Legend legend;
 final  List<Disclaimer> _disclaimers;
@override List<Disclaimer> get disclaimers {
  if (_disclaimers is EqualUnmodifiableListView) return _disclaimers;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_disclaimers);
}


/// Create a copy of HistoryYear
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$HistoryYearCopyWith<_HistoryYear> get copyWith => __$HistoryYearCopyWithImpl<_HistoryYear>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$HistoryYearToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _HistoryYear&&(identical(other.taxpayerId, taxpayerId) || other.taxpayerId == taxpayerId)&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&const DeepCollectionEquality().equals(other.months, _months)&&(identical(other.yearTotals, yearTotals) || other.yearTotals == yearTotals)&&(identical(other.legend, legend) || other.legend == legend)&&const DeepCollectionEquality().equals(other.disclaimers, _disclaimers));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,taxpayerId,fiscalYear,const DeepCollectionEquality().hash(_months),yearTotals,legend,const DeepCollectionEquality().hash(_disclaimers));
}

@override
String toString() {
    return 'HistoryYear(taxpayerId: $taxpayerId, fiscalYear: $fiscalYear, months: $months, yearTotals: $yearTotals, legend: $legend, disclaimers: $disclaimers)';
}


}

/// @nodoc
abstract mixin class _$HistoryYearCopyWith<$Res> implements $HistoryYearCopyWith<$Res> {
  factory _$HistoryYearCopyWith(_HistoryYear value, $Res Function(_HistoryYear) _then) = __$HistoryYearCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, List<HistoryMonth> months,@JsonKey(name: 'year_totals') HistoryYearTotals yearTotals, Legend legend, List<Disclaimer> disclaimers
});


@override $HistoryYearTotalsCopyWith<$Res> get yearTotals;

}
/// @nodoc
class __$HistoryYearCopyWithImpl<$Res>
    implements _$HistoryYearCopyWith<$Res> {
  __$HistoryYearCopyWithImpl(this._self, this._then);

  final _HistoryYear _self;
  final $Res Function(_HistoryYear) _then;

/// Create a copy of HistoryYear
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? taxpayerId = null,Object? fiscalYear = null,Object? months = null,Object? yearTotals = null,Object? legend = null,Object? disclaimers = null,}) {
  return _then(_HistoryYear(
taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,months: null == months ? _self._months : months // ignore: cast_nullable_to_non_nullable
as List<HistoryMonth>,yearTotals: null == yearTotals ? _self.yearTotals : yearTotals // ignore: cast_nullable_to_non_nullable
as HistoryYearTotals,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,disclaimers: null == disclaimers ? _self._disclaimers : disclaimers // ignore: cast_nullable_to_non_nullable
as List<Disclaimer>,
  ));
}

/// Create a copy of HistoryYear
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$HistoryYearTotalsCopyWith<$Res> get yearTotals {
  
  return $HistoryYearTotalsCopyWith<$Res>(_self.yearTotals, (value) {
    return _then(_self.copyWith(yearTotals: value));
  });
}
}

// dart format on
