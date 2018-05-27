// function should be called after async operation is finished
export function execFuncs(instance, execs) {
  // instance: enzyme mounted component
  // const execs = [
  //   {
  //     func: "componentWillMount",
  //     args: []
  //   }
  // ];
  return execs.reduce((prePromise, nextFunc) => {
    return prePromise.then(() => instance[nextFunc.func](...nextFunc.args));
  }, Promise.resolve());
}

export function execsToStr(execs) {
  // const execs = [
  //   {
  //     func: "componentWillMount",
  //     args: []
  //   }
  // ];
  const execList = execs.map(
    funcInfo => `${funcInfo.func}(${funcInfo.args.join(", ")})`
  );

  return execList.join(", ");
}

export function getDesc(componentName, testCase) {
  // const testCase = {
  //   execs: [
  //     {
  //       func: "onAddLocalFiles",
  //       args: []
  //     }
  //   ],
  //   state: {
  //     filterFileName: ""
  //   },
  //   calls: [
  //     {
  //       func: "onAddLocalFiles",
  //       count: 1
  //     }
  //   ]
  // }
  return `${componentName} should satisfy following by exec ${execsToStr(
    testCase.execs
  )}
        state=${JSON.stringify(testCase.state)}
        calls=${JSON.stringify(testCase.calls)} `;
}

export function verifyCalls(calls, stubs) {
  // const calls: [
  //   {
  //     func: "funcName",
  //     count: 1
  //   }
  // ];
  // const stubs = {
  //   funcName: jest.fn(),
  // };
  let err = null;
  calls.forEach(called => {
    if (stubs[called.func].mock.calls.length != called.count) {
      err = `InfoBar: ${called.func} should be called ${called.count} but ${
        stubs[called.func].mock.calls.length
      }`;
    }
  });
  return err;
}
