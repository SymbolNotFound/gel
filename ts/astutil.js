// Copyright (c) 2023 Symbol Not Found L.L.C.
// 
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//  
//      http://www.apache.org/licenses/LICENSE-2.0
//  
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
// 
// github:SymbolNotFound/ggdl/ts/gdl/astutil.js -- AST utility functions


// Returns a {line, col} position object for the beginning of a token.
function start(token) {
    return {
        line: token.line,
        col: token.col
    };
}

// Returns a {line, col} position for the end of a token.  NOTE: this assumes
// there are no newlines within the token (such as in space-capturing tokens).
// Albeit unlikely, avoid using this function on tokens that may contain `\n`.
function end(token) {
    return {
        line: token.line,
        col: token.col + token.text.length
    };
}

function newRole(name, start, end) {
  return {
    'type': 'role',
    'name': name,
    'start': start,
    'end': end
  };
}

function newRoleVar(name, start, end) {
  return {
    'type': 'role_var',
    'name': name,
    'start': start,
    'end': end
  };
}

function newUnaryOp(opName, child, start, end) {
  return {
    'type': 'unaryOp',
    'name': opName,
    'child': child,
    'start': start,
    'end': end
  };
}

function newBinaryOp(opName, left, right, start, end) {
  return {
    'type': 'binaryOp',
    'name': opName,
    'left': left,
    'right': right,
    'start': start,
    'end': end
  };
}

function newAction(actionType, roleName, action, start, end) {
  return {
    'type': actionType,
    'role': roleName,
    'action': action,
    'start': start,
    'end': end
  };
}

function newFunction(funcName, params, start, end) {
  return {
    'type': 'function',
    'name': funcName,
    'params': params,
    'start': start,
    'end': end
  };
}

function newInference(outcome, premises, start, end) {
  return {
    'type': 'infer',
    'outcome': outcome,
    'premises': premises,
    'start': start,
    'end': end
  };
}

function newGoal(agent, amount, start, end) {
  return {
    'type': 'goal',
    'agent': agent,
    'amount': amount,
    'start': start,
    'end': end
  };
}

function newTerminal(tok) {
  return {
    'type': 'terminal',
    'start': start(tok),
    'end': end(tok)
  };
}

function newName(name, start, end) {
  return {
    'type': 'name',
    'name': name,
    'start': start,
    'end': end
  };
}

function newVar(name, start, end) {
  return {
    'type': 'var',
    'name': name,
    'start': start,
    'end': end
  };
}

function newGroundRelation(relName, terms, start, end) {
  return {
    'type': "relation",
    'grounded': true,
    'name': relName,
    'terms': terms,
    'start': start,
    'end': end
  };
}
