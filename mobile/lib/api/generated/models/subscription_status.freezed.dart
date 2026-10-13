// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'subscription_status.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$SubscriptionStatus {

 AccessSummary get access;@JsonKey(includeIfNull: false) SubscriptionPlan? get plan;
/// Create a copy of SubscriptionStatus
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$SubscriptionStatusCopyWith<SubscriptionStatus> get copyWith => _$SubscriptionStatusCopyWithImpl<SubscriptionStatus>(this as SubscriptionStatus, _$identity);

  /// Serializes this SubscriptionStatus to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as SubscriptionStatus;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is SubscriptionStatus&&(identical(other.access, _this.access) || other.access == _this.access)&&(identical(other.plan, _this.plan) || other.plan == _this.plan));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as SubscriptionStatus;
  return Object.hash(runtimeType,_this.access,_this.plan);
}

@override
String toString() {
  final _this = this as SubscriptionStatus;
  return 'SubscriptionStatus(access: ${_this.access}, plan: ${_this.plan})';
}


}

/// @nodoc
abstract mixin class $SubscriptionStatusCopyWith<$Res>  {
  factory $SubscriptionStatusCopyWith(SubscriptionStatus value, $Res Function(SubscriptionStatus) _then) = _$SubscriptionStatusCopyWithImpl;
@useResult
$Res call({
 AccessSummary access,@JsonKey(includeIfNull: false) SubscriptionPlan? plan
});


$AccessSummaryCopyWith<$Res> get access;$SubscriptionPlanCopyWith<$Res>? get plan;

}
/// @nodoc
class _$SubscriptionStatusCopyWithImpl<$Res>
    implements $SubscriptionStatusCopyWith<$Res> {
  _$SubscriptionStatusCopyWithImpl(this._self, this._then);

  final SubscriptionStatus _self;
  final $Res Function(SubscriptionStatus) _then;

/// Create a copy of SubscriptionStatus
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? access = null,Object? plan = freezed,}) {
  return _then(SubscriptionStatus(
access: null == access ? _self.access : access // ignore: cast_nullable_to_non_nullable
as AccessSummary,plan: freezed == plan ? _self.plan : plan // ignore: cast_nullable_to_non_nullable
as SubscriptionPlan?,
  ));
}
/// Create a copy of SubscriptionStatus
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AccessSummaryCopyWith<$Res> get access {
  
  return $AccessSummaryCopyWith<$Res>(_self.access, (value) {
    return _then(_self.copyWith(access: value));
  });
}/// Create a copy of SubscriptionStatus
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$SubscriptionPlanCopyWith<$Res>? get plan {
    if (_self.plan == null) {
    return null;
  }

  return $SubscriptionPlanCopyWith<$Res>(_self.plan!, (value) {
    return _then(_self.copyWith(plan: value));
  });
}
}


/// Adds pattern-matching-related methods to [SubscriptionStatus].
extension SubscriptionStatusPatterns on SubscriptionStatus {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _SubscriptionStatus value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _SubscriptionStatus() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _SubscriptionStatus value)  $default,){
final _that = this;
switch (_that) {
case _SubscriptionStatus():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _SubscriptionStatus value)?  $default,){
final _that = this;
switch (_that) {
case _SubscriptionStatus() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( AccessSummary access, @JsonKey(includeIfNull: false)  SubscriptionPlan? plan)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _SubscriptionStatus() when $default != null:
return $default(_that.access,_that.plan);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( AccessSummary access, @JsonKey(includeIfNull: false)  SubscriptionPlan? plan)  $default,) {final _that = this;
switch (_that) {
case _SubscriptionStatus():
return $default(_that.access,_that.plan);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( AccessSummary access, @JsonKey(includeIfNull: false)  SubscriptionPlan? plan)?  $default,) {final _that = this;
switch (_that) {
case _SubscriptionStatus() when $default != null:
return $default(_that.access,_that.plan);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _SubscriptionStatus implements SubscriptionStatus {
  const _SubscriptionStatus({required this.access, @JsonKey(includeIfNull: false) this.plan});
  factory _SubscriptionStatus.fromJson(Map<String, dynamic> json) => _$SubscriptionStatusFromJson(json);

@override final  AccessSummary access;
@override@JsonKey(includeIfNull: false) final  SubscriptionPlan? plan;

/// Create a copy of SubscriptionStatus
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$SubscriptionStatusCopyWith<_SubscriptionStatus> get copyWith => __$SubscriptionStatusCopyWithImpl<_SubscriptionStatus>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$SubscriptionStatusToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _SubscriptionStatus&&(identical(other.access, access) || other.access == access)&&(identical(other.plan, plan) || other.plan == plan));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,access,plan);
}

@override
String toString() {
    return 'SubscriptionStatus(access: $access, plan: $plan)';
}


}

/// @nodoc
abstract mixin class _$SubscriptionStatusCopyWith<$Res> implements $SubscriptionStatusCopyWith<$Res> {
  factory _$SubscriptionStatusCopyWith(_SubscriptionStatus value, $Res Function(_SubscriptionStatus) _then) = __$SubscriptionStatusCopyWithImpl;
@override @useResult
$Res call({
 AccessSummary access,@JsonKey(includeIfNull: false) SubscriptionPlan? plan
});


@override $AccessSummaryCopyWith<$Res> get access;@override $SubscriptionPlanCopyWith<$Res>? get plan;

}
/// @nodoc
class __$SubscriptionStatusCopyWithImpl<$Res>
    implements _$SubscriptionStatusCopyWith<$Res> {
  __$SubscriptionStatusCopyWithImpl(this._self, this._then);

  final _SubscriptionStatus _self;
  final $Res Function(_SubscriptionStatus) _then;

/// Create a copy of SubscriptionStatus
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? access = null,Object? plan = freezed,}) {
  return _then(_SubscriptionStatus(
access: null == access ? _self.access : access // ignore: cast_nullable_to_non_nullable
as AccessSummary,plan: freezed == plan ? _self.plan : plan // ignore: cast_nullable_to_non_nullable
as SubscriptionPlan?,
  ));
}

/// Create a copy of SubscriptionStatus
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AccessSummaryCopyWith<$Res> get access {
  
  return $AccessSummaryCopyWith<$Res>(_self.access, (value) {
    return _then(_self.copyWith(access: value));
  });
}/// Create a copy of SubscriptionStatus
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$SubscriptionPlanCopyWith<$Res>? get plan {
    if (_self.plan == null) {
    return null;
  }

  return $SubscriptionPlanCopyWith<$Res>(_self.plan!, (value) {
    return _then(_self.copyWith(plan: value));
  });
}
}

// dart format on
