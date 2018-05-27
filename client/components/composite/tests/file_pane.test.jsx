jest.mock("../../../libs/api_share");
import React from "react";
import { FilePane } from "../file_pane";
import { mount } from "enzyme";
import * as mockApiShare from "../../../libs/api_share";
import { execFuncs, getDesc, verifyCalls } from "../../../tests/test_helper";
import valueEqual from "value-equal";

describe("FilePane", () => {
  const tests = [
    {
      init: {
        list: [{ Id: 0, PathLocal: "" }]
      },
      execs: [
        {
          func: "componentWillMount",
          args: []
        }
      ],
      state: {
        infos: [{ Id: 0, PathLocal: "" }],
        showDetailId: -1
      },
      calls: [
        {
          func: "onList",
          count: 2 // because componentWillMount will be callled twice
        },
        {
          func: "onOk",
          count: 2 // because componentWillMount will be callled twice
        }
      ]
    },
    {
      init: {
        list: [{ Id: 0, PathLocal: "" }, { Id: 1, PathLocal: "" }]
      },
      execs: [
        {
          func: "componentWillMount",
          args: []
        },
        {
          func: "onUpdateProgressImp",
          args: [0, "100%"]
        }
      ],
      state: {
        infos: [
          { Id: 0, PathLocal: "", progress: "100%" },
          { Id: 1, PathLocal: "" }
        ],
        showDetailId: -1
      }
    },
    {
      init: {
        list: []
      },
      execs: [
        {
          func: "componentWillMount",
          args: []
        },
        {
          func: "onToggleDetail",
          args: [0]
        }
      ],
      state: {
        infos: [],
        showDetailId: 0
      }
    },
    {
      init: {
        list: []
      },
      execs: [
        {
          func: "onToggleDetail",
          args: [0]
        },
        {
          func: "onToggleDetail",
          args: [0]
        }
      ],
      state: {
        infos: [],
        showDetailId: -1
      }
    }
  ];

  tests.forEach(testCase => {
    test(getDesc("FilePane", testCase), () => {
      // mock list()
      mockApiShare.__truncInfos();
      mockApiShare.__addInfos(testCase.init.list);

      const stubs = {
        onList: jest.fn(),
        onOk: jest.fn(),
        onError: jest.fn()
      };

      const stubWraps = {
        onListWrap: () => {
          stubs.onList();
          return mockApiShare.list();
        }
      };

      return new Promise((resolve, reject) => {
        const pane = mount(
          <FilePane
            onList={stubWraps.onListWrap}
            onOk={stubs.onOk}
            onError={stubs.onError}
          />
        );

        execFuncs(pane.instance(), testCase.execs).then(() => {
          pane.update();
          if (!valueEqual(pane.instance().state, testCase.state)) {
            return reject("FilePane: state not identical");
          }

          if (testCase.calls != null) {
            const err = verifyCalls(testCase.calls, stubs);
            if (err != null) {
              return reject(err);
            }
          }

          resolve();
        });
      });
    });
  });
});
