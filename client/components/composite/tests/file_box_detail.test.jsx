jest.mock("../../../libs/api_share");
import React from "react";
import { mount } from "enzyme";
import { FileBoxDetail, classDelYes, classDelNo } from "../file_box_detail";
import { execFuncs, getDesc, verifyCalls } from "../../../tests/test_helper";
import valueEqual from "value-equal";
import {
  del,
  publishId,
  shadowId,
  setDownLimit
} from "../../../libs/api_share";

describe("FileBoxDetail", () => {
  test("FileBoxDetail should show delete button by default, toggle using onComfirmDel and onCancelDel", () => {
    const box = mount(<FileBoxDetail />);
    expect(box.instance().state.showDelComfirm).toBe(false);
    box.instance().onComfirmDel();
    expect(box.instance().state.showDelComfirm).toBe(true);
    box.instance().onCancelDel();
    expect(box.instance().state.showDelComfirm).toBe(false);
  });
});

describe("FileBoxDetail", () => {
  const tests = [
    {
      init: {
        id: "0",
        name: "filename",
        size: "1B",
        modTime: 0,
        href: "href",
        downLimit: -1
      },
      execs: [
        {
          func: "onSetDownLimit",
          args: [3]
        }
      ],
      state: {
        downLimit: 3,
        showDelComfirm: false
      }
    },
    {
      init: {
        id: "0",
        name: "filename",
        size: "1B",
        modTime: 0,
        href: "href",
        downLimit: -1
      },
      execs: [
        {
          func: "onComfirmDel",
          args: []
        }
      ],
      state: {
        downLimit: -1,
        showDelComfirm: true
      }
    },
    {
      init: {
        id: "0",
        name: "filename",
        size: "1B",
        modTime: 0,
        href: "href",
        downLimit: -1
      },
      execs: [
        {
          func: "onComfirmDel",
          args: []
        },
        {
          func: "onCancelDel",
          args: []
        }
      ],
      state: {
        downLimit: -1,
        showDelComfirm: false
      }
    },
    {
      init: {
        id: "0",
        name: "filename",
        size: "1B",
        modTime: 0,
        href: "href",
        downLimit: -1
      },
      execs: [
        {
          func: "onResetLink",
          args: []
        }
      ],
      state: {
        downLimit: -1,
        showDelComfirm: false
      },
      calls: [
        {
          func: "onPublishId",
          count: 1
        },
        {
          func: "onOk",
          count: 1
        },
        {
          func: "onRefresh",
          count: 1
        }
      ]
    },
    {
      init: {
        id: "0",
        name: "filename",
        size: "1B",
        modTime: 0,
        href: "href",
        downLimit: -1
      },
      execs: [
        {
          func: "onShadowLink",
          args: []
        }
      ],
      state: {
        downLimit: -1,
        showDelComfirm: false
      },
      calls: [
        {
          func: "onShadowId",
          count: 1
        },
        {
          func: "onOk",
          count: 1
        },
        {
          func: "onRefresh",
          count: 1
        }
      ]
    },
    {
      init: {
        id: "0",
        name: "filename",
        size: "1B",
        modTime: 0,
        href: "href",
        downLimit: -1
      },
      execs: [
        {
          func: "onUpdateDownLimit",
          args: []
        }
      ],
      state: {
        downLimit: -1,
        showDelComfirm: false
      },
      calls: [
        {
          func: "onSetDownLimit",
          count: 1
        },
        {
          func: "onOk",
          count: 1
        },
        {
          func: "onRefresh",
          count: 1
        }
      ]
    },
    {
      init: {
        id: "0",
        name: "filename",
        size: "1B",
        modTime: 0,
        href: "href",
        downLimit: -1
      },
      execs: [
        {
          func: "onDelete",
          args: []
        }
      ],
      state: {
        downLimit: -1,
        showDelComfirm: false
      },
      calls: [
        {
          func: "onDel",
          count: 1
        },
        {
          func: "onOk",
          count: 1
        },
        {
          func: "onRefresh",
          count: 1
        }
      ]
    }
  ];

  tests.forEach(testCase => {
    test(getDesc("FileBoxDetail", testCase), () => {
      const stubs = {
        onOk: jest.fn(),
        onError: jest.fn(),
        onRefresh: jest.fn(),
        onDel: jest.fn(),
        onPublishId: jest.fn(),
        onShadowId: jest.fn(),
        onSetDownLimit: jest.fn()
      };

      const stubWraps = {
        onDel: () => {
          stubs.onDel();
          return del();
        },
        onPublishId: () => {
          stubs.onPublishId();
          return publishId();
        },
        onShadowId: () => {
          stubs.onShadowId();
          return shadowId();
        },
        onSetDownLimit: () => {
          stubs.onSetDownLimit();
          return setDownLimit();
        }
      };

      return new Promise((resolve, reject) => {
        const pane = mount(
          <FileBoxDetail
            id={testCase.init.id}
            name={testCase.init.name}
            size={testCase.init.size}
            modTime={testCase.init.modTime}
            href={testCase.init.href}
            downLimit={testCase.init.downLimit}
            onRefresh={stubs.onRefresh}
            onOk={stubs.onOk}
            onError={stubs.onError}
            onDel={stubWraps.onDel}
            onPublishId={stubWraps.onPublishId}
            onShadowId={stubWraps.onShadowId}
            onSetDownLimit={stubWraps.onSetDownLimit}
          />
        );

        execFuncs(pane.instance(), testCase.execs).then(() => {
          pane.update();
          if (!valueEqual(pane.instance().state, testCase.state)) {
            return reject("FileBoxDetail: state not identical");
          }

          if (testCase.calls != null) {
            const err = verifyCalls(testCase.calls, stubs);
            if (err != null) {
              return reject("FileBoxDetail: state not identical");
            }
          }

          resolve();
        });
      });
    });
  });
});
