// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'subscription_plan.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$SubscriptionPlan {

@JsonKey(name: 'product_id') String get productId; SubscriptionPlanPeriod get period; SubscriptionPlanStore get store;@JsonKey(name: 'expires_at') DateTime get expiresAt;@JsonKey(name: 'will_renew') bool get willRenew;@JsonKey(includeIfNull: false, name: 'management_url') String? get managementUrl;
/// Create a copy of SubscriptionPlan
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$SubscriptionPlanCopyWith<SubscriptionPlan> get copyWith => _$SubscriptionPlanCopyWithImpl<SubscriptionPlan>(this as SubscriptionPlan, _$identity);

  /// Serializes this SubscriptionPlan to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as SubscriptionPlan;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is SubscriptionPlan&&(identical(other.productId, _this.productId) || other.productId == _this.productId)&&(identical(other.period, _this.period) || other.period == _this.period)&&(identical(other.store, _this.store) || other.store == _this.store)&&(identical(other.expiresAt, _this.expiresAt) || other.expiresAt == _this.expiresAt)&&(identical(other.willRenew, _this.willRenew) || other.willRenew == _this.willRenew)&&(identical(other.managementUrl, _this.managementUrl) || other.managementUrl == _this.managementUrl));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as SubscriptionPlan;
  return Object.hash(runtimeType,_this.productId,_this.period,_this.store,_this.expiresAt,_this.willRenew,_this.managementUrl);
}

@override
String toString() {
  final _this = this as SubscriptionPlan;
  return 'SubscriptionPlan(productId: ${_this.productId}, period: ${_this.period}, store: ${_this.store}, expiresAt: ${_this.expiresAt}, willRenew: ${_this.willRenew}, managementUrl: ${_this.managementUrl})';
}


}

/// @nodoc
abstract mixin class $SubscriptionPlanCopyWith<$Res>  {
  factory $SubscriptionPlanCopyWith(SubscriptionPlan value, $Res Function(SubscriptionPlan) _then) = _$SubscriptionPlanCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'product_id') String productId, SubscriptionPlanPeriod period, SubscriptionPlanStore store,@JsonKey(name: 'expires_at') DateTime expiresAt,@JsonKey(name: 'will_renew') bool willRenew,@JsonKey(includeIfNull: false, name: 'management_url') String? managementUrl
});




}
/// @nodoc
class _$SubscriptionPlanCopyWithImpl<$Res>
    implements $SubscriptionPlanCopyWith<$Res> {
  _$SubscriptionPlanCopyWithImpl(this._self, this._then);

  final SubscriptionPlan _self;
  final $Res Function(SubscriptionPlan) _then;

/// Create a copy of SubscriptionPlan
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? productId = null,Object? period = null,Object? store = null,Object? expiresAt = null,Object? willRenew = null,Object? managementUrl = freezed,}) {
  return _then(SubscriptionPlan(
productId: null == productId ? _self.productId : productId // ignore: cast_nullable_to_non_nullable
as String,period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as SubscriptionPlanPeriod,store: null == store ? _self.store : store // ignore: cast_nullable_to_non_nullable
as SubscriptionPlanStore,expiresAt: null == expiresAt ? _self.expiresAt : expiresAt // ignore: cast_nullable_to_non_nullable
as DateTime,willRenew: null == willRenew ? _self.willRenew : willRenew // ignore: cast_nullable_to_non_nullable
as bool,managementUrl: freezed == managementUrl ? _self.managementUrl : managementUrl // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}

}


/// Adds pattern-matching-related methods to [SubscriptionPlan].
extension SubscriptionPlanPatterns on SubscriptionPlan {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _SubscriptionPlan value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _SubscriptionPlan() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _SubscriptionPlan value)  $default,){
final _that = this;
switch (_that) {
case _SubscriptionPlan():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _SubscriptionPlan value)?  $default,){
final _that = this;
switch (_that) {
case _SubscriptionPlan() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'product_id')  String productId,  SubscriptionPlanPeriod period,  SubscriptionPlanStore store, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'will_renew')  bool willRenew, @JsonKey(includeIfNull: false, name: 'management_url')  String? managementUrl)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _SubscriptionPlan() when $default != null:
return $default(_that.productId,_that.period,_that.store,_that.expiresAt,_that.willRenew,_that.managementUrl);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'product_id')  String productId,  SubscriptionPlanPeriod period,  SubscriptionPlanStore store, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'will_renew')  bool willRenew, @JsonKey(includeIfNull: false, name: 'management_url')  String? managementUrl)  $default,) {final _that = this;
switch (_that) {
case _SubscriptionPlan():
return $default(_that.productId,_that.period,_that.store,_that.expiresAt,_that.willRenew,_that.managementUrl);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'product_id')  String productId,  SubscriptionPlanPeriod period,  SubscriptionPlanStore store, @JsonKey(name: 'expires_at')  DateTime expiresAt, @JsonKey(name: 'will_renew')  bool willRenew, @JsonKey(includeIfNull: false, name: 'management_url')  String? managementUrl)?  $default,) {final _that = this;
switch (_that) {
case _SubscriptionPlan() when $default != null:
return $default(_that.productId,_that.period,_that.store,_that.expiresAt,_that.willRenew,_that.managementUrl);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _SubscriptionPlan implements SubscriptionPlan {
  const _SubscriptionPlan({@JsonKey(name: 'product_id') required this.productId, required this.period, required this.store, @JsonKey(name: 'expires_at') required this.expiresAt, @JsonKey(name: 'will_renew') required this.willRenew, @JsonKey(includeIfNull: false, name: 'management_url') this.managementUrl});
  factory _SubscriptionPlan.fromJson(Map<String, dynamic> json) => _$SubscriptionPlanFromJson(json);

@override@JsonKey(name: 'product_id') final  String productId;
@override final  SubscriptionPlanPeriod period;
@override final  SubscriptionPlanStore store;
@override@JsonKey(name: 'expires_at') final  DateTime expiresAt;
@override@JsonKey(name: 'will_renew') final  bool willRenew;
@override@JsonKey(includeIfNull: false, name: 'management_url') final  String? managementUrl;

/// Create a copy of SubscriptionPlan
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$SubscriptionPlanCopyWith<_SubscriptionPlan> get copyWith => __$SubscriptionPlanCopyWithImpl<_SubscriptionPlan>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$SubscriptionPlanToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _SubscriptionPlan&&(identical(other.productId, productId) || other.productId == productId)&&(identical(other.period, period) || other.period == period)&&(identical(other.store, store) || other.store == store)&&(identical(other.expiresAt, expiresAt) || other.expiresAt == expiresAt)&&(identical(other.willRenew, willRenew) || other.willRenew == willRenew)&&(identical(other.managementUrl, managementUrl) || other.managementUrl == managementUrl));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,productId,period,store,expiresAt,willRenew,managementUrl);
}

@override
String toString() {
    return 'SubscriptionPlan(productId: $productId, period: $period, store: $store, expiresAt: $expiresAt, willRenew: $willRenew, managementUrl: $managementUrl)';
}


}

/// @nodoc
abstract mixin class _$SubscriptionPlanCopyWith<$Res> implements $SubscriptionPlanCopyWith<$Res> {
  factory _$SubscriptionPlanCopyWith(_SubscriptionPlan value, $Res Function(_SubscriptionPlan) _then) = __$SubscriptionPlanCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'product_id') String productId, SubscriptionPlanPeriod period, SubscriptionPlanStore store,@JsonKey(name: 'expires_at') DateTime expiresAt,@JsonKey(name: 'will_renew') bool willRenew,@JsonKey(includeIfNull: false, name: 'management_url') String? managementUrl
});




}
/// @nodoc
class __$SubscriptionPlanCopyWithImpl<$Res>
    implements _$SubscriptionPlanCopyWith<$Res> {
  __$SubscriptionPlanCopyWithImpl(this._self, this._then);

  final _SubscriptionPlan _self;
  final $Res Function(_SubscriptionPlan) _then;

/// Create a copy of SubscriptionPlan
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? productId = null,Object? period = null,Object? store = null,Object? expiresAt = null,Object? willRenew = null,Object? managementUrl = freezed,}) {
  return _then(_SubscriptionPlan(
productId: null == productId ? _self.productId : productId // ignore: cast_nullable_to_non_nullable
as String,period: null == period ? _self.period : period // ignore: cast_nullable_to_non_nullable
as SubscriptionPlanPeriod,store: null == store ? _self.store : store // ignore: cast_nullable_to_non_nullable
as SubscriptionPlanStore,expiresAt: null == expiresAt ? _self.expiresAt : expiresAt // ignore: cast_nullable_to_non_nullable
as DateTime,willRenew: null == willRenew ? _self.willRenew : willRenew // ignore: cast_nullable_to_non_nullable
as bool,managementUrl: freezed == managementUrl ? _self.managementUrl : managementUrl // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}


}

// dart format on
