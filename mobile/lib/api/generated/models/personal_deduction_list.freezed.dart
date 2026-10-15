// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'personal_deduction_list.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$PersonalDeductionList {

@JsonKey(name: 'fiscal_year') FiscalYear get fiscalYear; List<PersonalDeduction> get items; List<DeductionCategoryTotal> get totals;
/// Create a copy of PersonalDeductionList
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$PersonalDeductionListCopyWith<PersonalDeductionList> get copyWith => _$PersonalDeductionListCopyWithImpl<PersonalDeductionList>(this as PersonalDeductionList, _$identity);

  /// Serializes this PersonalDeductionList to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as PersonalDeductionList;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is PersonalDeductionList&&(identical(other.fiscalYear, _this.fiscalYear) || other.fiscalYear == _this.fiscalYear)&&const DeepCollectionEquality().equals(other.items, _this.items)&&const DeepCollectionEquality().equals(other.totals, _this.totals));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as PersonalDeductionList;
  return Object.hash(runtimeType,_this.fiscalYear,const DeepCollectionEquality().hash(_this.items),const DeepCollectionEquality().hash(_this.totals));
}

@override
String toString() {
  final _this = this as PersonalDeductionList;
  return 'PersonalDeductionList(fiscalYear: ${_this.fiscalYear}, items: ${_this.items}, totals: ${_this.totals})';
}


}

/// @nodoc
abstract mixin class $PersonalDeductionListCopyWith<$Res>  {
  factory $PersonalDeductionListCopyWith(PersonalDeductionList value, $Res Function(PersonalDeductionList) _then) = _$PersonalDeductionListCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, List<PersonalDeduction> items, List<DeductionCategoryTotal> totals
});




}
/// @nodoc
class _$PersonalDeductionListCopyWithImpl<$Res>
    implements $PersonalDeductionListCopyWith<$Res> {
  _$PersonalDeductionListCopyWithImpl(this._self, this._then);

  final PersonalDeductionList _self;
  final $Res Function(PersonalDeductionList) _then;

/// Create a copy of PersonalDeductionList
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? fiscalYear = null,Object? items = null,Object? totals = null,}) {
  return _then(PersonalDeductionList(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,items: null == items ? _self.items : items // ignore: cast_nullable_to_non_nullable
as List<PersonalDeduction>,totals: null == totals ? _self.totals : totals // ignore: cast_nullable_to_non_nullable
as List<DeductionCategoryTotal>,
  ));
}

}


/// Adds pattern-matching-related methods to [PersonalDeductionList].
extension PersonalDeductionListPatterns on PersonalDeductionList {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _PersonalDeductionList value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _PersonalDeductionList() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _PersonalDeductionList value)  $default,){
final _that = this;
switch (_that) {
case _PersonalDeductionList():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _PersonalDeductionList value)?  $default,){
final _that = this;
switch (_that) {
case _PersonalDeductionList() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<PersonalDeduction> items,  List<DeductionCategoryTotal> totals)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _PersonalDeductionList() when $default != null:
return $default(_that.fiscalYear,_that.items,_that.totals);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<PersonalDeduction> items,  List<DeductionCategoryTotal> totals)  $default,) {final _that = this;
switch (_that) {
case _PersonalDeductionList():
return $default(_that.fiscalYear,_that.items,_that.totals);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'fiscal_year')  FiscalYear fiscalYear,  List<PersonalDeduction> items,  List<DeductionCategoryTotal> totals)?  $default,) {final _that = this;
switch (_that) {
case _PersonalDeductionList() when $default != null:
return $default(_that.fiscalYear,_that.items,_that.totals);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _PersonalDeductionList implements PersonalDeductionList {
  const _PersonalDeductionList({@JsonKey(name: 'fiscal_year') required this.fiscalYear, required  List<PersonalDeduction> items, required  List<DeductionCategoryTotal> totals}): _items = items,_totals = totals;
  factory _PersonalDeductionList.fromJson(Map<String, dynamic> json) => _$PersonalDeductionListFromJson(json);

@override@JsonKey(name: 'fiscal_year') final  FiscalYear fiscalYear;
 final  List<PersonalDeduction> _items;
@override List<PersonalDeduction> get items {
  if (_items is EqualUnmodifiableListView) return _items;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_items);
}

 final  List<DeductionCategoryTotal> _totals;
@override List<DeductionCategoryTotal> get totals {
  if (_totals is EqualUnmodifiableListView) return _totals;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_totals);
}


/// Create a copy of PersonalDeductionList
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$PersonalDeductionListCopyWith<_PersonalDeductionList> get copyWith => __$PersonalDeductionListCopyWithImpl<_PersonalDeductionList>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$PersonalDeductionListToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _PersonalDeductionList&&(identical(other.fiscalYear, fiscalYear) || other.fiscalYear == fiscalYear)&&const DeepCollectionEquality().equals(other.items, _items)&&const DeepCollectionEquality().equals(other.totals, _totals));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,fiscalYear,const DeepCollectionEquality().hash(_items),const DeepCollectionEquality().hash(_totals));
}

@override
String toString() {
    return 'PersonalDeductionList(fiscalYear: $fiscalYear, items: $items, totals: $totals)';
}


}

/// @nodoc
abstract mixin class _$PersonalDeductionListCopyWith<$Res> implements $PersonalDeductionListCopyWith<$Res> {
  factory _$PersonalDeductionListCopyWith(_PersonalDeductionList value, $Res Function(_PersonalDeductionList) _then) = __$PersonalDeductionListCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'fiscal_year') FiscalYear fiscalYear, List<PersonalDeduction> items, List<DeductionCategoryTotal> totals
});




}
/// @nodoc
class __$PersonalDeductionListCopyWithImpl<$Res>
    implements _$PersonalDeductionListCopyWith<$Res> {
  __$PersonalDeductionListCopyWithImpl(this._self, this._then);

  final _PersonalDeductionList _self;
  final $Res Function(_PersonalDeductionList) _then;

/// Create a copy of PersonalDeductionList
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? fiscalYear = null,Object? items = null,Object? totals = null,}) {
  return _then(_PersonalDeductionList(
fiscalYear: null == fiscalYear ? _self.fiscalYear : fiscalYear // ignore: cast_nullable_to_non_nullable
as FiscalYear,items: null == items ? _self._items : items // ignore: cast_nullable_to_non_nullable
as List<PersonalDeduction>,totals: null == totals ? _self._totals : totals // ignore: cast_nullable_to_non_nullable
as List<DeductionCategoryTotal>,
  ));
}


}

// dart format on
