// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'filed_payment_month.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$FiledPaymentMonth {

 int get month; Period get period;@JsonKey(name: 'in_window') bool get inWindow;@JsonKey(includeIfNull: true, name: 'filed_cents') NonNegativeCents? get filedCents;@JsonKey(includeIfNull: true, name: 'tax_input_cents') NonNegativeCents? get taxInputCents;@JsonKey(includeIfNull: true, name: 'app_estimate_cents') NonNegativeCents? get appEstimateCents;@JsonKey(name: 'app_estimate_incomplete') bool get appEstimateIncomplete;@JsonKey(includeIfNull: true, name: 'effective_cents') NonNegativeCents? get effectiveCents; FiledPaymentSource get source;
/// Create a copy of FiledPaymentMonth
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$FiledPaymentMonthCopyWith<FiledPaymentMonth> get copyWith => _$FiledPaymentMonthCopyWithImpl<FiledPaymentMonth>(this as FiledPaymentMonth, _$identity);

  /// Serializes this FiledPaymentMonth to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as FiledPaymentMonth;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is FiledPaymentMonth&&(identical(other.month, _this.month) || other.month == _this.month)&&(identical(other.period, _this.period) || other.period == _this.period)&&(identical(other.inWindow, _this.inWindow) || other.inWindow == _this.inWindow)&&(identical(other.filedCents, _this.filedCents) || other.filedCents == _this.filedCents)&&(identical(other.taxInputCents, _this.taxInputCents) || other.taxInputCents == _this.taxInputCents)&&(identical(other.appEstimateCents, _this.appEstimateCents) || other.appEstimateCents == _this.appEstimateCents)&&(identical(other.appEstimateIncomplete, _this.appEstimateIncomplete) || other.appEstimateIncomplete == _this.appEstimateIncomplete)&&(identical(other.effectiveCents, _this.effectiveCents) || other.effectiveCents == _this.effectiveCents)&&(identical(other.source, _this.source) || other.source == _this.source));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as FiledPaymentMonth;
  return Object.hash(runtimeType,_this.month,_this.period,_this.inWindow,_this.filedCents,_this.taxInputCents,_this.appEstimateCents,_this.appEstimateIncomplete,_this.effectiveCents,_this.source);
}

@override
String toString() {
  final _this = this as FiledPaymentMonth;
  return 'FiledPaymentMonth(month: ${_this.month}, period: ${_this.period}, inWindow: ${_this.inWindow}, filedCents: ${_this.filedCents}, taxInputCents: ${_this.taxInputCents}, appEstimateCents: ${_this.appEstimateCents}, appEstimateIncomplete: ${_this.appEstimateIncomplete}, effectiveCents: ${_this.effectiveCents}, source: ${_this.source})';
}


}

/// @nodoc
abstract mixin class $FiledPaymentMonthCopyWith<$Res>  {
  factory $FiledPaymentMonthCopyWith(FiledPaymentMonth value, $Res Function(FiledPaymentMonth) _then) = _$FiledPaymentMonthCopyWithImpl;
@useResult
$Res call({
 int month, Period period,@JsonKey(name: 'in_window') bool inWindow,@JsonKey(includeIfNull: true, name: 'filed_cents') NonNegativeCents? filedCents,@JsonKey(includeIfNull: true, name: 'tax_input_cents') NonNegativeCents? taxInputCents,@JsonKey(includeIfNull: true, name: 'app_estimate_cents') NonNegativeCents? appEstimateCents,@JsonKey(name: 'app_estimate_incomplete') bool appEstimateIncomplete,@JsonKey(includeIfNull: true, name: 'effective_cents') NonNegativeCents? effectiveCents, FiledPaymentSource source
});




}
/// @nodoc
class _$FiledPaymentMonthCopyWithImpl<$Res>
    implements $FiledPaymentMonthCopyWith<$Res> {
  _$FiledPaymentMonthCopyWithImpl(this._self, this._then);

  final FiledPaymentMonth _self;
  final $Res Function(FiledPaymentMonth) _then;

/// Create a copy of FiledPaymentMonth
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? month = null,Object? period = null,Object? inWindow = null,Object? filedCents = freezed,Object? taxInputCents = freezed,Object? appEstimateCents = freezed,Object? appEstimateIncomplete = null,Object? effectiveCents = freezed,Object? source = null,}) {
  return _then(FiledPaymentMonth(
month: null == month ? _self.month : month // ignore: cast_nullable_to_non_nullable
as int,period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,inWindow: null == inWindow ? _self.inWindow : inWindow // ignore: cast_nullable_to_non_nullable
as bool,filedCents: freezed == filedCents ? _self.filedCents : filedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,taxInputCents: freezed == taxInputCents ? _self.taxInputCents : taxInputCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,appEstimateCents: freezed == appEstimateCents ? _self.appEstimateCents : appEstimateCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,appEstimateIncomplete: null == appEstimateIncomplete ? _self.appEstimateIncomplete : appEstimateIncomplete // ignore: cast_nullable_to_non_nullable
as bool,effectiveCents: freezed == effectiveCents ? _self.effectiveCents : effectiveCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,source: null == source ? _self.source : source // ignore: cast_nullable_to_non_nullable
as FiledPaymentSource,
  ));
}

}


/// Adds pattern-matching-related methods to [FiledPaymentMonth].
extension FiledPaymentMonthPatterns on FiledPaymentMonth {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _FiledPaymentMonth value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _FiledPaymentMonth() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _FiledPaymentMonth value)  $default,){
final _that = this;
switch (_that) {
case _FiledPaymentMonth():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _FiledPaymentMonth value)?  $default,){
final _that = this;
switch (_that) {
case _FiledPaymentMonth() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( int month,  Period period, @JsonKey(name: 'in_window')  bool inWindow, @JsonKey(includeIfNull: true, name: 'filed_cents')  NonNegativeCents? filedCents, @JsonKey(includeIfNull: true, name: 'tax_input_cents')  NonNegativeCents? taxInputCents, @JsonKey(includeIfNull: true, name: 'app_estimate_cents')  NonNegativeCents? appEstimateCents, @JsonKey(name: 'app_estimate_incomplete')  bool appEstimateIncomplete, @JsonKey(includeIfNull: true, name: 'effective_cents')  NonNegativeCents? effectiveCents,  FiledPaymentSource source)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _FiledPaymentMonth() when $default != null:
return $default(_that.month,_that.period,_that.inWindow,_that.filedCents,_that.taxInputCents,_that.appEstimateCents,_that.appEstimateIncomplete,_that.effectiveCents,_that.source);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( int month,  Period period, @JsonKey(name: 'in_window')  bool inWindow, @JsonKey(includeIfNull: true, name: 'filed_cents')  NonNegativeCents? filedCents, @JsonKey(includeIfNull: true, name: 'tax_input_cents')  NonNegativeCents? taxInputCents, @JsonKey(includeIfNull: true, name: 'app_estimate_cents')  NonNegativeCents? appEstimateCents, @JsonKey(name: 'app_estimate_incomplete')  bool appEstimateIncomplete, @JsonKey(includeIfNull: true, name: 'effective_cents')  NonNegativeCents? effectiveCents,  FiledPaymentSource source)  $default,) {final _that = this;
switch (_that) {
case _FiledPaymentMonth():
return $default(_that.month,_that.period,_that.inWindow,_that.filedCents,_that.taxInputCents,_that.appEstimateCents,_that.appEstimateIncomplete,_that.effectiveCents,_that.source);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( int month,  Period period, @JsonKey(name: 'in_window')  bool inWindow, @JsonKey(includeIfNull: true, name: 'filed_cents')  NonNegativeCents? filedCents, @JsonKey(includeIfNull: true, name: 'tax_input_cents')  NonNegativeCents? taxInputCents, @JsonKey(includeIfNull: true, name: 'app_estimate_cents')  NonNegativeCents? appEstimateCents, @JsonKey(name: 'app_estimate_incomplete')  bool appEstimateIncomplete, @JsonKey(includeIfNull: true, name: 'effective_cents')  NonNegativeCents? effectiveCents,  FiledPaymentSource source)?  $default,) {final _that = this;
switch (_that) {
case _FiledPaymentMonth() when $default != null:
return $default(_that.month,_that.period,_that.inWindow,_that.filedCents,_that.taxInputCents,_that.appEstimateCents,_that.appEstimateIncomplete,_that.effectiveCents,_that.source);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _FiledPaymentMonth implements FiledPaymentMonth {
  const _FiledPaymentMonth({required this.month, required this.period, @JsonKey(name: 'in_window') required this.inWindow, @JsonKey(includeIfNull: true, name: 'filed_cents') required this.filedCents, @JsonKey(includeIfNull: true, name: 'tax_input_cents') required this.taxInputCents, @JsonKey(includeIfNull: true, name: 'app_estimate_cents') required this.appEstimateCents, @JsonKey(name: 'app_estimate_incomplete') required this.appEstimateIncomplete, @JsonKey(includeIfNull: true, name: 'effective_cents') required this.effectiveCents, required this.source});
  factory _FiledPaymentMonth.fromJson(Map<String, dynamic> json) => _$FiledPaymentMonthFromJson(json);

@override final  int month;
@override final  Period period;
@override@JsonKey(name: 'in_window') final  bool inWindow;
@override@JsonKey(includeIfNull: true, name: 'filed_cents') final  NonNegativeCents? filedCents;
@override@JsonKey(includeIfNull: true, name: 'tax_input_cents') final  NonNegativeCents? taxInputCents;
@override@JsonKey(includeIfNull: true, name: 'app_estimate_cents') final  NonNegativeCents? appEstimateCents;
@override@JsonKey(name: 'app_estimate_incomplete') final  bool appEstimateIncomplete;
@override@JsonKey(includeIfNull: true, name: 'effective_cents') final  NonNegativeCents? effectiveCents;
@override final  FiledPaymentSource source;

/// Create a copy of FiledPaymentMonth
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$FiledPaymentMonthCopyWith<_FiledPaymentMonth> get copyWith => __$FiledPaymentMonthCopyWithImpl<_FiledPaymentMonth>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$FiledPaymentMonthToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _FiledPaymentMonth&&(identical(other.month, month) || other.month == month)&&(identical(other.period, period) || other.period == period)&&(identical(other.inWindow, inWindow) || other.inWindow == inWindow)&&(identical(other.filedCents, filedCents) || other.filedCents == filedCents)&&(identical(other.taxInputCents, taxInputCents) || other.taxInputCents == taxInputCents)&&(identical(other.appEstimateCents, appEstimateCents) || other.appEstimateCents == appEstimateCents)&&(identical(other.appEstimateIncomplete, appEstimateIncomplete) || other.appEstimateIncomplete == appEstimateIncomplete)&&(identical(other.effectiveCents, effectiveCents) || other.effectiveCents == effectiveCents)&&(identical(other.source, source) || other.source == source));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,month,period,inWindow,filedCents,taxInputCents,appEstimateCents,appEstimateIncomplete,effectiveCents,source);
}

@override
String toString() {
    return 'FiledPaymentMonth(month: $month, period: $period, inWindow: $inWindow, filedCents: $filedCents, taxInputCents: $taxInputCents, appEstimateCents: $appEstimateCents, appEstimateIncomplete: $appEstimateIncomplete, effectiveCents: $effectiveCents, source: $source)';
}


}

/// @nodoc
abstract mixin class _$FiledPaymentMonthCopyWith<$Res> implements $FiledPaymentMonthCopyWith<$Res> {
  factory _$FiledPaymentMonthCopyWith(_FiledPaymentMonth value, $Res Function(_FiledPaymentMonth) _then) = __$FiledPaymentMonthCopyWithImpl;
@override @useResult
$Res call({
 int month, Period period,@JsonKey(name: 'in_window') bool inWindow,@JsonKey(includeIfNull: true, name: 'filed_cents') NonNegativeCents? filedCents,@JsonKey(includeIfNull: true, name: 'tax_input_cents') NonNegativeCents? taxInputCents,@JsonKey(includeIfNull: true, name: 'app_estimate_cents') NonNegativeCents? appEstimateCents,@JsonKey(name: 'app_estimate_incomplete') bool appEstimateIncomplete,@JsonKey(includeIfNull: true, name: 'effective_cents') NonNegativeCents? effectiveCents, FiledPaymentSource source
});




}
/// @nodoc
class __$FiledPaymentMonthCopyWithImpl<$Res>
    implements _$FiledPaymentMonthCopyWith<$Res> {
  __$FiledPaymentMonthCopyWithImpl(this._self, this._then);

  final _FiledPaymentMonth _self;
  final $Res Function(_FiledPaymentMonth) _then;

/// Create a copy of FiledPaymentMonth
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? month = null,Object? period = null,Object? inWindow = null,Object? filedCents = freezed,Object? taxInputCents = freezed,Object? appEstimateCents = freezed,Object? appEstimateIncomplete = null,Object? effectiveCents = freezed,Object? source = null,}) {
  return _then(_FiledPaymentMonth(
month: null == month ? _self.month : month // ignore: cast_nullable_to_non_nullable
as int,period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,inWindow: null == inWindow ? _self.inWindow : inWindow // ignore: cast_nullable_to_non_nullable
as bool,filedCents: freezed == filedCents ? _self.filedCents : filedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,taxInputCents: freezed == taxInputCents ? _self.taxInputCents : taxInputCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,appEstimateCents: freezed == appEstimateCents ? _self.appEstimateCents : appEstimateCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,appEstimateIncomplete: null == appEstimateIncomplete ? _self.appEstimateIncomplete : appEstimateIncomplete // ignore: cast_nullable_to_non_nullable
as bool,effectiveCents: freezed == effectiveCents ? _self.effectiveCents : effectiveCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,source: null == source ? _self.source : source // ignore: cast_nullable_to_non_nullable
as FiledPaymentSource,
  ));
}


}

// dart format on
