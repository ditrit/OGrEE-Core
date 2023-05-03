import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:ogree_app/pages/select_page.dart';
import 'package:syncfusion_flutter_datepicker/datepicker.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

class SelectDate extends StatefulWidget {
  const SelectDate({super.key});
  @override
  State<SelectDate> createState() => _SelectDateState();
}

// Sample datasets
const List<String> datasetOptions = [
  '19/12/2022 - Jeu ABCDEF',
  '18/12/2022 - Jeu JKLMNO',
  '17/12/2022 - Jeu UVWXYZ'
];

class _SelectDateState extends State<SelectDate> with TickerProviderStateMixin {
  late TabController _tabController;
  String? _dataset = datasetOptions.first;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
  }

  @override
  Widget build(BuildContext context) {
    final localeMsg = AppLocalizations.of(context)!;

    return Column(
      children: [
        Text(
          localeMsg.whatDate,
          style: Theme.of(context).textTheme.headlineLarge,
        ),
        const SizedBox(height: 20),
        Card(
          child: Container(
              alignment: Alignment.center,
              child: Column(
                children: [
                  Align(
                    alignment: Alignment.center,
                    child: TabBar(
                      controller: _tabController,
                      labelPadding: const EdgeInsets.only(left: 20, right: 20),
                      labelColor: Colors.black,
                      unselectedLabelColor: Colors.grey,
                      labelStyle: TextStyle(
                          fontSize: 14,
                          fontFamily: GoogleFonts.inter().fontFamily),
                      unselectedLabelStyle: TextStyle(
                          fontSize: 14,
                          fontFamily: GoogleFonts.inter().fontFamily),
                      isScrollable: true,
                      indicatorSize: TabBarIndicatorSize.label,
                      tabs: [
                        Tab(
                          text: localeMsg.pickDate,
                          icon: Icon(Icons.calendar_month),
                        ),
                        Tab(
                          text: localeMsg.openLastDataset,
                          icon: Icon(Icons.timelapse),
                        ),
                        Tab(
                          text: localeMsg.openSavedDataser,
                          icon: Icon(Icons.calendar_view_day),
                        ),
                      ],
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.only(left: 20),
                    height: 350,
                    width: double.maxFinite,
                    child: TabBarView(
                      controller: _tabController,
                      children: [
                        const DatePicker(),
                        Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Text(
                              localeMsg.useLastDataSet,
                              style: Theme.of(context).textTheme.headlineMedium,
                            ),
                            const SizedBox(height: 32),
                            SizedBox(
                              width: 500.0,
                              height: 70.0,
                              child: OutlinedButton(
                                onPressed: () {},
                                autofocus: true,
                                child: Text(
                                  'Données mises à jour le 19/12/2022 à 19h45',
                                  style: GoogleFonts.inter(
                                    fontSize: 17,
                                  ),
                                ),
                              ),
                            )
                          ],
                        ),
                        Center(
                          child: SizedBox(
                            width: 500,
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: datasetOptions
                                  .map((dataset) => RadioListTile<String>(
                                        title: Text(dataset),
                                        value: dataset,
                                        groupValue: _dataset,
                                        onChanged: (String? value) {
                                          setState(() {
                                            _dataset = value;
                                          });
                                        },
                                      ))
                                  .toList(),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              )),
        ),
      ],
    );
  }
}

class DatePicker extends StatefulWidget {
  const DatePicker({
    Key? key,
  }) : super(key: key);

  @override
  State<DatePicker> createState() => _DatePickerState();
}

class _DatePickerState extends State<DatePicker> {
  String _selectedDate = '';
  String _dateCount = '';
  String _range = '';
  String _rangeCount = '';

  /// The method for [DateRangePickerSelectionChanged] callback, which will be
  /// called whenever a selection changed on the date picker widget.
  void _onSelectionChanged(DateRangePickerSelectionChangedArgs args) {
    /// The argument value will return the changed date as [DateTime] when the
    /// widget [SfDateRangeSelectionMode] set as single.
    ///
    /// The argument value will return the changed dates as [List<DateTime>]
    /// when the widget [SfDateRangeSelectionMode] set as multiple.
    ///
    /// The argument value will return the changed range as [PickerDateRange]
    /// when the widget [SfDateRangeSelectionMode] set as range.
    ///
    /// The argument value will return the changed ranges as
    /// [List<PickerDateRange] when the widget [SfDateRangeSelectionMode] set as
    /// multi range.
    setState(() {
      if (args.value is PickerDateRange) {
        _range = '${DateFormat('dd/MM/yyyy').format(args.value.startDate)} -'
            // ignore: lines_longer_than_80_chars
            ' ${DateFormat('dd/MM/yyyy').format(args.value.endDate ?? args.value.startDate)}';
        SelectPage.of(context)!.selectedDate = _range;
      } else if (args.value is DateTime) {
        _selectedDate = args.value.toString();
      } else if (args.value is List<DateTime>) {
        _dateCount = args.value.length.toString();
      } else {
        _rangeCount = args.value.length.toString();
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    print(DateTime.now());
    return Center(
        child: SizedBox(
      width: 700,
      height: 700,
      child: Container(
        padding: const EdgeInsets.fromLTRB(5, 30, 5, 5),
        child: SfDateRangePicker(
          onSelectionChanged: _onSelectionChanged,
          selectionMode: DateRangePickerSelectionMode.range,
          enableMultiView: MediaQuery.of(context).size.width > 700,
          headerStyle:
              const DateRangePickerHeaderStyle(textAlign: TextAlign.center),
          initialSelectedRange: PickerDateRange(
              // DateTime.now().subtract(const Duration(days: 4)),
              // DateTime.now().add(const Duration(days: 3))
              DateTime.now(),
              DateTime.now()),
        ),
      ),
    ));
  }
}
