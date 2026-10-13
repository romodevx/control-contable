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

 Period get period; HistoryMonthStatus get status;@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents get totalPayableRoundedCents;@JsonKey(name: 'changed_since_viewed') bool get changedSinceViewed; TaxResult? get iva; TaxResult? get isr;@JsonKey(name: 'last_viewed_at') DateTime? get lastViewedAt;
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
  return identical(this, other) || (other.runtimeType == runtimeType&&other is HistoryMonth&&(identical(other.period, _this.period) || other.period == _this.period)&&(identical(other.status, _this.status) || other.status == _this.status)&&(identical(other.totalPayableRoundedCents, _this.totalPayableRoundedCents) || other.totalPayableRoundedCents == _this.totalPayableRoundedCents)&&(identical(other.changedSinceViewed, _this.changedSinceViewed) || other.changedSinceViewed == _this.changedSinceViewed)&&(identical(other.iva, _this.iva) || other.iva == _this.iva)&&(identical(other.isr, _this.isr) || other.isr == _this.isr)&&(identical(other.lastViewedAt, _this.lastViewedAt) || other.lastViewedAt == _this.lastViewedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as HistoryMonth;
  return Object.hash(runtimeType,_this.period,_this.status,_this.totalPayableRoundedCents,_this.changedSinceViewed,_this.iva,_this.isr,_this.lastViewedAt);
}

@override
String toString() {
  final _this = this as HistoryMonth;
  return 'HistoryMonth(period: ${_this.period}, status: ${_this.status}, totalPayableRoundedCents: ${_this.totalPayableRoundedCents}, changedSinceViewed: ${_this.changedSinceViewed}, iva: ${_this.iva}, isr: ${_this.isr}, lastViewedAt: ${_this.lastViewedAt})';
}


}

/// @nodoc
abstract mixin class $HistoryMonthCopyWith<$Res>  {
  factory $HistoryMonthCopyWith(HistoryMonth value, $Res Function(HistoryMonth) _then) = _$HistoryMonthCopyWithImpl;
@useResult
$Res call({
 Period period, HistoryMonthStatus status,@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents totalPayableRoundedCents,@JsonKey(name: 'changed_since_viewed') bool changedSinceViewed, TaxResult? iva, TaxResult? isr,@JsonKey(name: 'last_viewed_at') DateTime? lastViewedAt
});


$TaxResultCopyWith<$Res>? get iva;$TaxResultCopyWith<$Res>? get isr;

}
/// @nodoc
class _$HistoryMonthCopyWithImpl<$Res>
    implements $HistoryMonthCopyWith<$Res> {
  _$HistoryMonthCopyWithImpl(this._self, this._then);

  final HistoryMonth _self;
  final $Res Function(HistoryMonth) _then;

/// Create a copy of HistoryMonth
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? period = null,Object? status = null,Object? totalPayableRoundedCents = null,Object? changedSinceViewed = null,Object? iva = freezed,Object? isr = freezed,Object? lastViewedAt = freezed,}) {
  return _then(HistoryMonth(
period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as HistoryMonthStatus,totalPayableRoundedCents: null == totalPayableRoundedCents ? _self.totalPayableRoundedCents : totalPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,changedSinceViewed: null == changedSinceViewed ? _self.changedSinceViewed : changedSinceViewed // ignore: cast_nullable_to_non_nullable
as bool,iva: freezed == iva ? _self.iva : iva // ignore: cast_nullable_to_non_nullable
as TaxResult?,isr: freezed == isr ? _self.isr : isr // ignore: cast_nullable_to_non_nullable
as TaxResult?,lastViewedAt: freezed == lastViewedAt ? _self.lastViewedAt : lastViewedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}
/// Create a copy of HistoryMonth
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TaxResultCopyWith<$Res>? get iva {
    if (_self.iva == null) {
    return null;
  }

  return $TaxResultCopyWith<$Res>(_self.iva!, (value) {
    return _then(_self.copyWith(iva: value));
  });
}/// Create a copy of HistoryMonth
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Period period,  HistoryMonthStatus status, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents, @JsonKey(name: 'changed_since_viewed')  bool changedSinceViewed,  TaxResult? iva,  TaxResult? isr, @JsonKey(name: 'last_viewed_at')  DateTime? lastViewedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _HistoryMonth() when $default != null:
return $default(_that.period,_that.status,_that.totalPayableRoundedCents,_that.changedSinceViewed,_that.iva,_that.isr,_that.lastViewedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Period period,  HistoryMonthStatus status, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents, @JsonKey(name: 'changed_since_viewed')  bool changedSinceViewed,  TaxResult? iva,  TaxResult? isr, @JsonKey(name: 'last_viewed_at')  DateTime? lastViewedAt)  $default,) {final _that = this;
switch (_that) {
case _HistoryMonth():
return $default(_that.period,_that.status,_that.totalPayableRoundedCents,_that.changedSinceViewed,_that.iva,_that.isr,_that.lastViewedAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Period period,  HistoryMonthStatus status, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents, @JsonKey(name: 'changed_since_viewed')  bool changedSinceViewed,  TaxResult? iva,  TaxResult? isr, @JsonKey(name: 'last_viewed_at')  DateTime? lastViewedAt)?  $default,) {final _that = this;
switch (_that) {
case _HistoryMonth() when $default != null:
return $default(_that.period,_that.status,_that.totalPayableRoundedCents,_that.changedSinceViewed,_that.iva,_that.isr,_that.lastViewedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _HistoryMonth implements HistoryMonth {
  const _HistoryMonth({required this.period, required this.status, @JsonKey(name: 'total_payable_rounded_cents') required this.totalPayableRoundedCents, @JsonKey(name: 'changed_since_viewed') required this.changedSinceViewed, this.iva, this.isr, @JsonKey(name: 'last_viewed_at') this.lastViewedAt});
  factory _HistoryMonth.fromJson(Map<String, dynamic> json) => _$HistoryMonthFromJson(json);

@override final  Period period;
@override final  HistoryMonthStatus status;
@override@JsonKey(name: 'total_payable_rounded_cents') final  NonNegativeCents totalPayableRoundedCents;
@override@JsonKey(name: 'changed_since_viewed') final  bool changedSinceViewed;
@override final  TaxResult? iva;
@override final  TaxResult? isr;
@override@JsonKey(name: 'last_viewed_at') final  DateTime? lastViewedAt;

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
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _HistoryMonth&&(identical(other.period, period) || other.period == period)&&(identical(other.status, status) || other.status == status)&&(identical(other.totalPayableRoundedCents, totalPayableRoundedCents) || other.totalPayableRoundedCents == totalPayableRoundedCents)&&(identical(other.changedSinceViewed, changedSinceViewed) || other.changedSinceViewed == changedSinceViewed)&&(identical(other.iva, iva) || other.iva == iva)&&(identical(other.isr, isr) || other.isr == isr)&&(identical(other.lastViewedAt, lastViewedAt) || other.lastViewedAt == lastViewedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,period,status,totalPayableRoundedCents,changedSinceViewed,iva,isr,lastViewedAt);
}

@override
String toString() {
    return 'HistoryMonth(period: $period, status: $status, totalPayableRoundedCents: $totalPayableRoundedCents, changedSinceViewed: $changedSinceViewed, iva: $iva, isr: $isr, lastViewedAt: $lastViewedAt)';
}


}

/// @nodoc
abstract mixin class _$HistoryMonthCopyWith<$Res> implements $HistoryMonthCopyWith<$Res> {
  factory _$HistoryMonthCopyWith(_HistoryMonth value, $Res Function(_HistoryMonth) _then) = __$HistoryMonthCopyWithImpl;
@override @useResult
$Res call({
 Period period, HistoryMonthStatus status,@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents totalPayableRoundedCents,@JsonKey(name: 'changed_since_viewed') bool changedSinceViewed, TaxResult? iva, TaxResult? isr,@JsonKey(name: 'last_viewed_at') DateTime? lastViewedAt
});


@override $TaxResultCopyWith<$Res>? get iva;@override $TaxResultCopyWith<$Res>? get isr;

}
/// @nodoc
class __$HistoryMonthCopyWithImpl<$Res>
    implements _$HistoryMonthCopyWith<$Res> {
  __$HistoryMonthCopyWithImpl(this._self, this._then);

  final _HistoryMonth _self;
  final $Res Function(_HistoryMonth) _then;

/// Create a copy of HistoryMonth
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? period = null,Object? status = null,Object? totalPayableRoundedCents = null,Object? changedSinceViewed = null,Object? iva = freezed,Object? isr = freezed,Object? lastViewedAt = freezed,}) {
  return _then(_HistoryMonth(
period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as Period,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as HistoryMonthStatus,totalPayableRoundedCents: null == totalPayableRoundedCents ? _self.totalPayableRoundedCents : totalPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,changedSinceViewed: null == changedSinceViewed ? _self.changedSinceViewed : changedSinceViewed // ignore: cast_nullable_to_non_nullable
as bool,iva: freezed == iva ? _self.iva : iva // ignore: cast_nullable_to_non_nullable
as TaxResult?,isr: freezed == isr ? _self.isr : isr // ignore: cast_nullable_to_non_nullable
as TaxResult?,lastViewedAt: freezed == lastViewedAt ? _self.lastViewedAt : lastViewedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}

/// Create a copy of HistoryMonth
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TaxResultCopyWith<$Res>? get iva {
    if (_self.iva == null) {
    return null;
  }

  return $TaxResultCopyWith<$Res>(_self.iva!, (value) {
    return _then(_self.copyWith(iva: value));
  });
}/// Create a copy of HistoryMonth
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
}
}

// dart format on
