import { PanelPlugin } from '@grafana/data';
import { TablePanel } from './TablePanel';
import { CustomFieldConfig, Options } from './types';
import { tableMigrationHandler, tablePanelChangedHandler } from './migrations';
import { TableCellDisplayMode } from '@grafana/ui';

export const plugin = new PanelPlugin<Options, CustomFieldConfig>(TablePanel)
  .setPanelChangeHandler(tablePanelChangedHandler)
  .setMigrationHandler(tableMigrationHandler)
  .setNoPadding()
  .useFieldConfig({
    useCustomConfig: builder => {
      builder
        .addNumberInput({
          path: 'width',
          name: '列宽',
          settings: {
            placeholder: 'auto',
            min: 20,
            max: 300,
          },
          shouldApply: () => true,
        })
        .addRadio({
          path: 'align',
          name: '列对齐方式',
          settings: {
            options: [
              { label: 'auto', value: null },
              { label: 'left', value: 'left' },
              { label: 'center', value: 'center' },
              { label: 'right', value: 'right' },
            ],
          },
          defaultValue: null,
        })
        .addSelect({
          path: 'displayMode',
          name: '单元显示模式',
          description: 'Color text, background, show as gauge, etc',
          settings: {
            options: [
              { value: TableCellDisplayMode.Auto, label: 'Auto' },
              { value: TableCellDisplayMode.ColorText, label: 'Color text' },
              { value: TableCellDisplayMode.ColorBackground, label: 'Color background' },
              { value: TableCellDisplayMode.GradientGauge, label: 'Gradient gauge' },
              { value: TableCellDisplayMode.LcdGauge, label: 'LCD gauge' },
              { value: TableCellDisplayMode.BasicGauge, label: 'Basic gauge' },
              { value: TableCellDisplayMode.JSONView, label: 'JSON View' },
              { value: TableCellDisplayMode.Image, label: 'Image' },
            ],
          },
        })
        .addBooleanSwitch({
          path: 'filterable',
          name: 'Column filter',
          description: 'Enables/disables field filters in table',
          defaultValue: false,
        });
    },
  })
  .setPanelOptions(builder => {
    builder.addBooleanSwitch({
      path: 'showHeader',
      name: 'Show header',
      description: "To display table's header or not to display",
      defaultValue: true,
    });
  });
