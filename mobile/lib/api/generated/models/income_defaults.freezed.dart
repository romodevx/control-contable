// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'income_defaults.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$IncomeDefaults {

@JsonKey(name: 'iva_rates_allowed') List<IvaRate> get ivaRatesAllowed;@JsonKey(name: 'suggested_iva_rate') IvaRate get suggestedIvaRate;@JsonKey(name: 'last_used') IncomeLastUsed get lastUsed;/// The 625 activity to preselect; `null` outside 625 or when it is ambiguous.
@JsonKey(includeIfNull: true, name: 'platform_activity') PlatformActivity? get platformActivity;
/// Create a copy of IncomeDefaults
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$IncomeDefaultsCopyWith<IncomeDefaults> get copyWith => _$IncomeDefaultsCopyWithImpl<IncomeDefaults>(this as IncomeDefaults, _$identity);

  /// Serializes this IncomeDefaults to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as IncomeDefaults;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is IncomeDefaults&&const DeepCollectionEquality().equals(other.ivaRatesAllowed, _this.ivaRatesAllowed)&&(identical(other.suggestedIvaRate, _this.suggestedIvaRate) || other.suggestedIvaRate == _this.suggestedIvaRate)&&(identical(other.lastUsed, _this.lastUsed) || other.lastUsed == _this.lastUsed)&&(identical(other.platformActivity, _this.platformActivity) || other.platformActivity == _this.platformActivity));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as IncomeDefaults;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.ivaRatesAllowed),_this.suggestedIvaRate,_this.lastUsed,_this.platformActivity);
}

@override
String toString() {
  final _this = this as IncomeDefaults;
  return 'IncomeDefaults(ivaRatesAllowed: ${_this.ivaRatesAllowed}, suggestedIvaRate: ${_this.suggestedIvaRate}, lastUsed: ${_this.lastUsed}, platformActivity: ${_this.platformActivity})';
}


}

/// @nodoc
abstract mixin class $IncomeDefaultsCopyWith<$Res>  {
  factory $IncomeDefaultsCopyWith(IncomeDefaults value, $Res Function(IncomeDefaults) _then) = _$IncomeDefaultsCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'iva_rates_allowed') List<IvaRate> ivaRatesAllowed,@JsonKey(name: 'suggested_iva_rate') IvaRate suggestedIvaRate,@JsonKey(name: 'last_used') IncomeLastUsed lastUsed,@JsonKey(includeIfNull: true, name: 'platform_activity') PlatformActivity? platformActivity
});


$IncomeLastUsedCopyWith<$Res> get lastUsed;

}
/// @nodoc
class _$IncomeDefaultsCopyWithImpl<$Res>
    implements $IncomeDefaultsCopyWith<$Res> {
  _$IncomeDefaultsCopyWithImpl(this._self, this._then);

  final IncomeDefaults _self;
  final $Res Function(IncomeDefaults) _then;

/// Create a copy of IncomeDefaults
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? ivaRatesAllowed = null,Object? suggestedIvaRate = null,Object? lastUsed = null,Object? platformActivity = freezed,}) {
  return _then(IncomeDefaults(
ivaRatesAllowed: null == ivaRatesAllowed ? _self.ivaRatesAllowed : ivaRatesAllowed // ignore: cast_nullable_to_non_nullable
as List<IvaRate>,suggestedIvaRate: null == suggestedIvaRate ? _self.suggestedIvaRate : suggestedIvaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,lastUsed: null == lastUsed ? _self.lastUsed : lastUsed // ignore: cast_nullable_to_non_nullable
as IncomeLastUsed,platformActivity: freezed == platformActivity ? _self.platformActivity : platformActivity // ignore: cast_nullable_to_non_nullable
as PlatformActivity?,
  ));
}
/// Create a copy of IncomeDefaults
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$IncomeLastUsedCopyWith<$Res> get lastUsed {
  
  return $IncomeLastUsedCopyWith<$Res>(_self.lastUsed, (value) {
    return _then(_self.copyWith(lastUsed: value));
  });
}
}


/// Adds pattern-matching-related methods to [IncomeDefaults].
extension IncomeDefaultsPatterns on IncomeDefaults {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _IncomeDefaults value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _IncomeDefaults() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _IncomeDefaults value)  $default,){
final _that = this;
switch (_that) {
case _IncomeDefaults():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _IncomeDefaults value)?  $default,){
final _that = this;
switch (_that) {
case _IncomeDefaults() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'iva_rates_allowed')  List<IvaRate> ivaRatesAllowed, @JsonKey(name: 'suggested_iva_rate')  IvaRate suggestedIvaRate, @JsonKey(name: 'last_used')  IncomeLastUsed lastUsed, @JsonKey(includeIfNull: true, name: 'platform_activity')  PlatformActivity? platformActivity)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _IncomeDefaults() when $default != null:
return $default(_that.ivaRatesAllowed,_that.suggestedIvaRate,_that.lastUsed,_that.platformActivity);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'iva_rates_allowed')  List<IvaRate> ivaRatesAllowed, @JsonKey(name: 'suggested_iva_rate')  IvaRate suggestedIvaRate, @JsonKey(name: 'last_used')  IncomeLastUsed lastUsed, @JsonKey(includeIfNull: true, name: 'platform_activity')  PlatformActivity? platformActivity)  $default,) {final _that = this;
switch (_that) {
case _IncomeDefaults():
return $default(_that.ivaRatesAllowed,_that.suggestedIvaRate,_that.lastUsed,_that.platformActivity);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'iva_rates_allowed')  List<IvaRate> ivaRatesAllowed, @JsonKey(name: 'suggested_iva_rate')  IvaRate suggestedIvaRate, @JsonKey(name: 'last_used')  IncomeLastUsed lastUsed, @JsonKey(includeIfNull: true, name: 'platform_activity')  PlatformActivity? platformActivity)?  $default,) {final _that = this;
switch (_that) {
case _IncomeDefaults() when $default != null:
return $default(_that.ivaRatesAllowed,_that.suggestedIvaRate,_that.lastUsed,_that.platformActivity);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _IncomeDefaults implements IncomeDefaults {
  const _IncomeDefaults({@JsonKey(name: 'iva_rates_allowed') required  List<IvaRate> ivaRatesAllowed, @JsonKey(name: 'suggested_iva_rate') required this.suggestedIvaRate, @JsonKey(name: 'last_used') required this.lastUsed, @JsonKey(includeIfNull: true, name: 'platform_activity') required this.platformActivity}): _ivaRatesAllowed = ivaRatesAllowed;
  factory _IncomeDefaults.fromJson(Map<String, dynamic> json) => _$IncomeDefaultsFromJson(json);

 final  List<IvaRate> _ivaRatesAllowed;
@override@JsonKey(name: 'iva_rates_allowed') List<IvaRate> get ivaRatesAllowed {
  if (_ivaRatesAllowed is EqualUnmodifiableListView) return _ivaRatesAllowed;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_ivaRatesAllowed);
}

@override@JsonKey(name: 'suggested_iva_rate') final  IvaRate suggestedIvaRate;
@override@JsonKey(name: 'last_used') final  IncomeLastUsed lastUsed;
/// The 625 activity to preselect; `null` outside 625 or when it is ambiguous.
@override@JsonKey(includeIfNull: true, name: 'platform_activity') final  PlatformActivity? platformActivity;

/// Create a copy of IncomeDefaults
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$IncomeDefaultsCopyWith<_IncomeDefaults> get copyWith => __$IncomeDefaultsCopyWithImpl<_IncomeDefaults>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$IncomeDefaultsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _IncomeDefaults&&const DeepCollectionEquality().equals(other.ivaRatesAllowed, _ivaRatesAllowed)&&(identical(other.suggestedIvaRate, suggestedIvaRate) || other.suggestedIvaRate == suggestedIvaRate)&&(identical(other.lastUsed, lastUsed) || other.lastUsed == lastUsed)&&(identical(other.platformActivity, platformActivity) || other.platformActivity == platformActivity));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_ivaRatesAllowed),suggestedIvaRate,lastUsed,platformActivity);
}

@override
String toString() {
    return 'IncomeDefaults(ivaRatesAllowed: $ivaRatesAllowed, suggestedIvaRate: $suggestedIvaRate, lastUsed: $lastUsed, platformActivity: $platformActivity)';
}


}

/// @nodoc
abstract mixin class _$IncomeDefaultsCopyWith<$Res> implements $IncomeDefaultsCopyWith<$Res> {
  factory _$IncomeDefaultsCopyWith(_IncomeDefaults value, $Res Function(_IncomeDefaults) _then) = __$IncomeDefaultsCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'iva_rates_allowed') List<IvaRate> ivaRatesAllowed,@JsonKey(name: 'suggested_iva_rate') IvaRate suggestedIvaRate,@JsonKey(name: 'last_used') IncomeLastUsed lastUsed,@JsonKey(includeIfNull: true, name: 'platform_activity') PlatformActivity? platformActivity
});


@override $IncomeLastUsedCopyWith<$Res> get lastUsed;

}
/// @nodoc
class __$IncomeDefaultsCopyWithImpl<$Res>
    implements _$IncomeDefaultsCopyWith<$Res> {
  __$IncomeDefaultsCopyWithImpl(this._self, this._then);

  final _IncomeDefaults _self;
  final $Res Function(_IncomeDefaults) _then;

/// Create a copy of IncomeDefaults
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? ivaRatesAllowed = null,Object? suggestedIvaRate = null,Object? lastUsed = null,Object? platformActivity = freezed,}) {
  return _then(_IncomeDefaults(
ivaRatesAllowed: null == ivaRatesAllowed ? _self._ivaRatesAllowed : ivaRatesAllowed // ignore: cast_nullable_to_non_nullable
as List<IvaRate>,suggestedIvaRate: null == suggestedIvaRate ? _self.suggestedIvaRate : suggestedIvaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,lastUsed: null == lastUsed ? _self.lastUsed : lastUsed // ignore: cast_nullable_to_non_nullable
as IncomeLastUsed,platformActivity: freezed == platformActivity ? _self.platformActivity : platformActivity // ignore: cast_nullable_to_non_nullable
as PlatformActivity?,
  ));
}

/// Create a copy of IncomeDefaults
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$IncomeLastUsedCopyWith<$Res> get lastUsed {
  
  return $IncomeLastUsedCopyWith<$Res>(_self.lastUsed, (value) {
    return _then(_self.copyWith(lastUsed: value));
  });
}
}

// dart format on
