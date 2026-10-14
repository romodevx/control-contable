// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'expense_page.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ExpensePage {

 List<Expense> get items;/// Opaque cursor of the next page; `null` on the last page.
@JsonKey(includeIfNull: true, name: 'next_cursor') String? get nextCursor; PeriodTotals get totals; Legend get legend;
/// Create a copy of ExpensePage
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ExpensePageCopyWith<ExpensePage> get copyWith => _$ExpensePageCopyWithImpl<ExpensePage>(this as ExpensePage, _$identity);

  /// Serializes this ExpensePage to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ExpensePage;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ExpensePage&&const DeepCollectionEquality().equals(other.items, _this.items)&&(identical(other.nextCursor, _this.nextCursor) || other.nextCursor == _this.nextCursor)&&(identical(other.totals, _this.totals) || other.totals == _this.totals)&&(identical(other.legend, _this.legend) || other.legend == _this.legend));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ExpensePage;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.items),_this.nextCursor,_this.totals,_this.legend);
}

@override
String toString() {
  final _this = this as ExpensePage;
  return 'ExpensePage(items: ${_this.items}, nextCursor: ${_this.nextCursor}, totals: ${_this.totals}, legend: ${_this.legend})';
}


}

/// @nodoc
abstract mixin class $ExpensePageCopyWith<$Res>  {
  factory $ExpensePageCopyWith(ExpensePage value, $Res Function(ExpensePage) _then) = _$ExpensePageCopyWithImpl;
@useResult
$Res call({
 List<Expense> items,@JsonKey(includeIfNull: true, name: 'next_cursor') String? nextCursor, PeriodTotals totals, Legend legend
});


$PeriodTotalsCopyWith<$Res> get totals;

}
/// @nodoc
class _$ExpensePageCopyWithImpl<$Res>
    implements $ExpensePageCopyWith<$Res> {
  _$ExpensePageCopyWithImpl(this._self, this._then);

  final ExpensePage _self;
  final $Res Function(ExpensePage) _then;

/// Create a copy of ExpensePage
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? items = null,Object? nextCursor = freezed,Object? totals = null,Object? legend = null,}) {
  return _then(ExpensePage(
items: null == items ? _self.items : items // ignore: cast_nullable_to_non_nullable
as List<Expense>,nextCursor: freezed == nextCursor ? _self.nextCursor : nextCursor // ignore: cast_nullable_to_non_nullable
as String?,totals: null == totals ? _self.totals : totals // ignore: cast_nullable_to_non_nullable
as PeriodTotals,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,
  ));
}
/// Create a copy of ExpensePage
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$PeriodTotalsCopyWith<$Res> get totals {
  
  return $PeriodTotalsCopyWith<$Res>(_self.totals, (value) {
    return _then(_self.copyWith(totals: value));
  });
}
}


/// Adds pattern-matching-related methods to [ExpensePage].
extension ExpensePagePatterns on ExpensePage {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ExpensePage value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ExpensePage() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ExpensePage value)  $default,){
final _that = this;
switch (_that) {
case _ExpensePage():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ExpensePage value)?  $default,){
final _that = this;
switch (_that) {
case _ExpensePage() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<Expense> items, @JsonKey(includeIfNull: true, name: 'next_cursor')  String? nextCursor,  PeriodTotals totals,  Legend legend)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ExpensePage() when $default != null:
return $default(_that.items,_that.nextCursor,_that.totals,_that.legend);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<Expense> items, @JsonKey(includeIfNull: true, name: 'next_cursor')  String? nextCursor,  PeriodTotals totals,  Legend legend)  $default,) {final _that = this;
switch (_that) {
case _ExpensePage():
return $default(_that.items,_that.nextCursor,_that.totals,_that.legend);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<Expense> items, @JsonKey(includeIfNull: true, name: 'next_cursor')  String? nextCursor,  PeriodTotals totals,  Legend legend)?  $default,) {final _that = this;
switch (_that) {
case _ExpensePage() when $default != null:
return $default(_that.items,_that.nextCursor,_that.totals,_that.legend);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ExpensePage implements ExpensePage {
  const _ExpensePage({required  List<Expense> items, @JsonKey(includeIfNull: true, name: 'next_cursor') required this.nextCursor, required this.totals, required this.legend}): _items = items;
  factory _ExpensePage.fromJson(Map<String, dynamic> json) => _$ExpensePageFromJson(json);

 final  List<Expense> _items;
@override List<Expense> get items {
  if (_items is EqualUnmodifiableListView) return _items;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_items);
}

/// Opaque cursor of the next page; `null` on the last page.
@override@JsonKey(includeIfNull: true, name: 'next_cursor') final  String? nextCursor;
@override final  PeriodTotals totals;
@override final  Legend legend;

/// Create a copy of ExpensePage
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ExpensePageCopyWith<_ExpensePage> get copyWith => __$ExpensePageCopyWithImpl<_ExpensePage>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ExpensePageToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ExpensePage&&const DeepCollectionEquality().equals(other.items, _items)&&(identical(other.nextCursor, nextCursor) || other.nextCursor == nextCursor)&&(identical(other.totals, totals) || other.totals == totals)&&(identical(other.legend, legend) || other.legend == legend));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_items),nextCursor,totals,legend);
}

@override
String toString() {
    return 'ExpensePage(items: $items, nextCursor: $nextCursor, totals: $totals, legend: $legend)';
}


}

/// @nodoc
abstract mixin class _$ExpensePageCopyWith<$Res> implements $ExpensePageCopyWith<$Res> {
  factory _$ExpensePageCopyWith(_ExpensePage value, $Res Function(_ExpensePage) _then) = __$ExpensePageCopyWithImpl;
@override @useResult
$Res call({
 List<Expense> items,@JsonKey(includeIfNull: true, name: 'next_cursor') String? nextCursor, PeriodTotals totals, Legend legend
});


@override $PeriodTotalsCopyWith<$Res> get totals;

}
/// @nodoc
class __$ExpensePageCopyWithImpl<$Res>
    implements _$ExpensePageCopyWith<$Res> {
  __$ExpensePageCopyWithImpl(this._self, this._then);

  final _ExpensePage _self;
  final $Res Function(_ExpensePage) _then;

/// Create a copy of ExpensePage
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? items = null,Object? nextCursor = freezed,Object? totals = null,Object? legend = null,}) {
  return _then(_ExpensePage(
items: null == items ? _self._items : items // ignore: cast_nullable_to_non_nullable
as List<Expense>,nextCursor: freezed == nextCursor ? _self.nextCursor : nextCursor // ignore: cast_nullable_to_non_nullable
as String?,totals: null == totals ? _self.totals : totals // ignore: cast_nullable_to_non_nullable
as PeriodTotals,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,
  ));
}

/// Create a copy of ExpensePage
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$PeriodTotalsCopyWith<$Res> get totals {
  
  return $PeriodTotalsCopyWith<$Res>(_self.totals, (value) {
    return _then(_self.copyWith(totals: value));
  });
}
}

// dart format on
