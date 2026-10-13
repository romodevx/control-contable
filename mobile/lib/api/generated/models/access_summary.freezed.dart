// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'access_summary.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AccessSummary {

 AccessSummaryAccess get access;@JsonKey(name: 'trial_ends_at') DateTime? get trialEndsAt;@JsonKey(name: 'incomes_remaining') int? get incomesRemaining;@JsonKey(name: 'expenses_remaining') int? get expensesRemaining;
/// Create a copy of AccessSummary
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AccessSummaryCopyWith<AccessSummary> get copyWith => _$AccessSummaryCopyWithImpl<AccessSummary>(this as AccessSummary, _$identity);

  /// Serializes this AccessSummary to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AccessSummary;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AccessSummary&&(identical(other.access, _this.access) || other.access == _this.access)&&(identical(other.trialEndsAt, _this.trialEndsAt) || other.trialEndsAt == _this.trialEndsAt)&&(identical(other.incomesRemaining, _this.incomesRemaining) || other.incomesRemaining == _this.incomesRemaining)&&(identical(other.expensesRemaining, _this.expensesRemaining) || other.expensesRemaining == _this.expensesRemaining));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AccessSummary;
  return Object.hash(runtimeType,_this.access,_this.trialEndsAt,_this.incomesRemaining,_this.expensesRemaining);
}

@override
String toString() {
  final _this = this as AccessSummary;
  return 'AccessSummary(access: ${_this.access}, trialEndsAt: ${_this.trialEndsAt}, incomesRemaining: ${_this.incomesRemaining}, expensesRemaining: ${_this.expensesRemaining})';
}


}

/// @nodoc
abstract mixin class $AccessSummaryCopyWith<$Res>  {
  factory $AccessSummaryCopyWith(AccessSummary value, $Res Function(AccessSummary) _then) = _$AccessSummaryCopyWithImpl;
@useResult
$Res call({
 AccessSummaryAccess access,@JsonKey(name: 'trial_ends_at') DateTime? trialEndsAt,@JsonKey(name: 'incomes_remaining') int? incomesRemaining,@JsonKey(name: 'expenses_remaining') int? expensesRemaining
});




}
/// @nodoc
class _$AccessSummaryCopyWithImpl<$Res>
    implements $AccessSummaryCopyWith<$Res> {
  _$AccessSummaryCopyWithImpl(this._self, this._then);

  final AccessSummary _self;
  final $Res Function(AccessSummary) _then;

/// Create a copy of AccessSummary
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? access = null,Object? trialEndsAt = freezed,Object? incomesRemaining = freezed,Object? expensesRemaining = freezed,}) {
  return _then(AccessSummary(
access: null == access ? _self.access : access // ignore: cast_nullable_to_non_nullable
as AccessSummaryAccess,trialEndsAt: freezed == trialEndsAt ? _self.trialEndsAt : trialEndsAt // ignore: cast_nullable_to_non_nullable
as DateTime?,incomesRemaining: freezed == incomesRemaining ? _self.incomesRemaining : incomesRemaining // ignore: cast_nullable_to_non_nullable
as int?,expensesRemaining: freezed == expensesRemaining ? _self.expensesRemaining : expensesRemaining // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}

}


/// Adds pattern-matching-related methods to [AccessSummary].
extension AccessSummaryPatterns on AccessSummary {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AccessSummary value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AccessSummary() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AccessSummary value)  $default,){
final _that = this;
switch (_that) {
case _AccessSummary():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AccessSummary value)?  $default,){
final _that = this;
switch (_that) {
case _AccessSummary() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( AccessSummaryAccess access, @JsonKey(name: 'trial_ends_at')  DateTime? trialEndsAt, @JsonKey(name: 'incomes_remaining')  int? incomesRemaining, @JsonKey(name: 'expenses_remaining')  int? expensesRemaining)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AccessSummary() when $default != null:
return $default(_that.access,_that.trialEndsAt,_that.incomesRemaining,_that.expensesRemaining);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( AccessSummaryAccess access, @JsonKey(name: 'trial_ends_at')  DateTime? trialEndsAt, @JsonKey(name: 'incomes_remaining')  int? incomesRemaining, @JsonKey(name: 'expenses_remaining')  int? expensesRemaining)  $default,) {final _that = this;
switch (_that) {
case _AccessSummary():
return $default(_that.access,_that.trialEndsAt,_that.incomesRemaining,_that.expensesRemaining);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( AccessSummaryAccess access, @JsonKey(name: 'trial_ends_at')  DateTime? trialEndsAt, @JsonKey(name: 'incomes_remaining')  int? incomesRemaining, @JsonKey(name: 'expenses_remaining')  int? expensesRemaining)?  $default,) {final _that = this;
switch (_that) {
case _AccessSummary() when $default != null:
return $default(_that.access,_that.trialEndsAt,_that.incomesRemaining,_that.expensesRemaining);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AccessSummary implements AccessSummary {
  const _AccessSummary({required this.access, @JsonKey(name: 'trial_ends_at') this.trialEndsAt, @JsonKey(name: 'incomes_remaining') this.incomesRemaining, @JsonKey(name: 'expenses_remaining') this.expensesRemaining});
  factory _AccessSummary.fromJson(Map<String, dynamic> json) => _$AccessSummaryFromJson(json);

@override final  AccessSummaryAccess access;
@override@JsonKey(name: 'trial_ends_at') final  DateTime? trialEndsAt;
@override@JsonKey(name: 'incomes_remaining') final  int? incomesRemaining;
@override@JsonKey(name: 'expenses_remaining') final  int? expensesRemaining;

/// Create a copy of AccessSummary
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AccessSummaryCopyWith<_AccessSummary> get copyWith => __$AccessSummaryCopyWithImpl<_AccessSummary>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AccessSummaryToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AccessSummary&&(identical(other.access, access) || other.access == access)&&(identical(other.trialEndsAt, trialEndsAt) || other.trialEndsAt == trialEndsAt)&&(identical(other.incomesRemaining, incomesRemaining) || other.incomesRemaining == incomesRemaining)&&(identical(other.expensesRemaining, expensesRemaining) || other.expensesRemaining == expensesRemaining));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,access,trialEndsAt,incomesRemaining,expensesRemaining);
}

@override
String toString() {
    return 'AccessSummary(access: $access, trialEndsAt: $trialEndsAt, incomesRemaining: $incomesRemaining, expensesRemaining: $expensesRemaining)';
}


}

/// @nodoc
abstract mixin class _$AccessSummaryCopyWith<$Res> implements $AccessSummaryCopyWith<$Res> {
  factory _$AccessSummaryCopyWith(_AccessSummary value, $Res Function(_AccessSummary) _then) = __$AccessSummaryCopyWithImpl;
@override @useResult
$Res call({
 AccessSummaryAccess access,@JsonKey(name: 'trial_ends_at') DateTime? trialEndsAt,@JsonKey(name: 'incomes_remaining') int? incomesRemaining,@JsonKey(name: 'expenses_remaining') int? expensesRemaining
});




}
/// @nodoc
class __$AccessSummaryCopyWithImpl<$Res>
    implements _$AccessSummaryCopyWith<$Res> {
  __$AccessSummaryCopyWithImpl(this._self, this._then);

  final _AccessSummary _self;
  final $Res Function(_AccessSummary) _then;

/// Create a copy of AccessSummary
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? access = null,Object? trialEndsAt = freezed,Object? incomesRemaining = freezed,Object? expensesRemaining = freezed,}) {
  return _then(_AccessSummary(
access: null == access ? _self.access : access // ignore: cast_nullable_to_non_nullable
as AccessSummaryAccess,trialEndsAt: freezed == trialEndsAt ? _self.trialEndsAt : trialEndsAt // ignore: cast_nullable_to_non_nullable
as DateTime?,incomesRemaining: freezed == incomesRemaining ? _self.incomesRemaining : incomesRemaining // ignore: cast_nullable_to_non_nullable
as int?,expensesRemaining: freezed == expensesRemaining ? _self.expensesRemaining : expensesRemaining // ignore: cast_nullable_to_non_nullable
as int?,
  ));
}


}

// dart format on
