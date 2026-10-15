// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'deduction_category_total.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$DeductionCategoryTotal {

 DeductionCategory get category;@JsonKey(name: 'registered_cents') NonNegativeCents get registeredCents;@JsonKey(name: 'excluded_cents') NonNegativeCents get excludedCents; int get count;
/// Create a copy of DeductionCategoryTotal
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$DeductionCategoryTotalCopyWith<DeductionCategoryTotal> get copyWith => _$DeductionCategoryTotalCopyWithImpl<DeductionCategoryTotal>(this as DeductionCategoryTotal, _$identity);

  /// Serializes this DeductionCategoryTotal to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as DeductionCategoryTotal;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is DeductionCategoryTotal&&(identical(other.category, _this.category) || other.category == _this.category)&&(identical(other.registeredCents, _this.registeredCents) || other.registeredCents == _this.registeredCents)&&(identical(other.excludedCents, _this.excludedCents) || other.excludedCents == _this.excludedCents)&&(identical(other.count, _this.count) || other.count == _this.count));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as DeductionCategoryTotal;
  return Object.hash(runtimeType,_this.category,_this.registeredCents,_this.excludedCents,_this.count);
}

@override
String toString() {
  final _this = this as DeductionCategoryTotal;
  return 'DeductionCategoryTotal(category: ${_this.category}, registeredCents: ${_this.registeredCents}, excludedCents: ${_this.excludedCents}, count: ${_this.count})';
}


}

/// @nodoc
abstract mixin class $DeductionCategoryTotalCopyWith<$Res>  {
  factory $DeductionCategoryTotalCopyWith(DeductionCategoryTotal value, $Res Function(DeductionCategoryTotal) _then) = _$DeductionCategoryTotalCopyWithImpl;
@useResult
$Res call({
 DeductionCategory category,@JsonKey(name: 'registered_cents') NonNegativeCents registeredCents,@JsonKey(name: 'excluded_cents') NonNegativeCents excludedCents, int count
});




}
/// @nodoc
class _$DeductionCategoryTotalCopyWithImpl<$Res>
    implements $DeductionCategoryTotalCopyWith<$Res> {
  _$DeductionCategoryTotalCopyWithImpl(this._self, this._then);

  final DeductionCategoryTotal _self;
  final $Res Function(DeductionCategoryTotal) _then;

/// Create a copy of DeductionCategoryTotal
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? category = null,Object? registeredCents = null,Object? excludedCents = null,Object? count = null,}) {
  return _then(DeductionCategoryTotal(
category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as DeductionCategory,registeredCents: null == registeredCents ? _self.registeredCents : registeredCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,excludedCents: null == excludedCents ? _self.excludedCents : excludedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,count: null == count ? _self.count : count // ignore: cast_nullable_to_non_nullable
as int,
  ));
}

}


/// Adds pattern-matching-related methods to [DeductionCategoryTotal].
extension DeductionCategoryTotalPatterns on DeductionCategoryTotal {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _DeductionCategoryTotal value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _DeductionCategoryTotal() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _DeductionCategoryTotal value)  $default,){
final _that = this;
switch (_that) {
case _DeductionCategoryTotal():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _DeductionCategoryTotal value)?  $default,){
final _that = this;
switch (_that) {
case _DeductionCategoryTotal() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( DeductionCategory category, @JsonKey(name: 'registered_cents')  NonNegativeCents registeredCents, @JsonKey(name: 'excluded_cents')  NonNegativeCents excludedCents,  int count)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _DeductionCategoryTotal() when $default != null:
return $default(_that.category,_that.registeredCents,_that.excludedCents,_that.count);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( DeductionCategory category, @JsonKey(name: 'registered_cents')  NonNegativeCents registeredCents, @JsonKey(name: 'excluded_cents')  NonNegativeCents excludedCents,  int count)  $default,) {final _that = this;
switch (_that) {
case _DeductionCategoryTotal():
return $default(_that.category,_that.registeredCents,_that.excludedCents,_that.count);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( DeductionCategory category, @JsonKey(name: 'registered_cents')  NonNegativeCents registeredCents, @JsonKey(name: 'excluded_cents')  NonNegativeCents excludedCents,  int count)?  $default,) {final _that = this;
switch (_that) {
case _DeductionCategoryTotal() when $default != null:
return $default(_that.category,_that.registeredCents,_that.excludedCents,_that.count);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _DeductionCategoryTotal implements DeductionCategoryTotal {
  const _DeductionCategoryTotal({required this.category, @JsonKey(name: 'registered_cents') required this.registeredCents, @JsonKey(name: 'excluded_cents') required this.excludedCents, required this.count});
  factory _DeductionCategoryTotal.fromJson(Map<String, dynamic> json) => _$DeductionCategoryTotalFromJson(json);

@override final  DeductionCategory category;
@override@JsonKey(name: 'registered_cents') final  NonNegativeCents registeredCents;
@override@JsonKey(name: 'excluded_cents') final  NonNegativeCents excludedCents;
@override final  int count;

/// Create a copy of DeductionCategoryTotal
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$DeductionCategoryTotalCopyWith<_DeductionCategoryTotal> get copyWith => __$DeductionCategoryTotalCopyWithImpl<_DeductionCategoryTotal>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$DeductionCategoryTotalToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _DeductionCategoryTotal&&(identical(other.category, category) || other.category == category)&&(identical(other.registeredCents, registeredCents) || other.registeredCents == registeredCents)&&(identical(other.excludedCents, excludedCents) || other.excludedCents == excludedCents)&&(identical(other.count, count) || other.count == count));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,category,registeredCents,excludedCents,count);
}

@override
String toString() {
    return 'DeductionCategoryTotal(category: $category, registeredCents: $registeredCents, excludedCents: $excludedCents, count: $count)';
}


}

/// @nodoc
abstract mixin class _$DeductionCategoryTotalCopyWith<$Res> implements $DeductionCategoryTotalCopyWith<$Res> {
  factory _$DeductionCategoryTotalCopyWith(_DeductionCategoryTotal value, $Res Function(_DeductionCategoryTotal) _then) = __$DeductionCategoryTotalCopyWithImpl;
@override @useResult
$Res call({
 DeductionCategory category,@JsonKey(name: 'registered_cents') NonNegativeCents registeredCents,@JsonKey(name: 'excluded_cents') NonNegativeCents excludedCents, int count
});




}
/// @nodoc
class __$DeductionCategoryTotalCopyWithImpl<$Res>
    implements _$DeductionCategoryTotalCopyWith<$Res> {
  __$DeductionCategoryTotalCopyWithImpl(this._self, this._then);

  final _DeductionCategoryTotal _self;
  final $Res Function(_DeductionCategoryTotal) _then;

/// Create a copy of DeductionCategoryTotal
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? category = null,Object? registeredCents = null,Object? excludedCents = null,Object? count = null,}) {
  return _then(_DeductionCategoryTotal(
category: null == category ? _self.category : category // ignore: cast_nullable_to_non_nullable
as DeductionCategory,registeredCents: null == registeredCents ? _self.registeredCents : registeredCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,excludedCents: null == excludedCents ? _self.excludedCents : excludedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,count: null == count ? _self.count : count // ignore: cast_nullable_to_non_nullable
as int,
  ));
}


}

// dart format on
