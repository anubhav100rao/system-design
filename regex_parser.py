# Regex engine

SPECIAL_SYMBOLS = {'.', '*', '+', '?', '|'}
GROUP_SYMBOLS = {'[', ']'}
ESCAPE_CHARACTER = '\\'

def tokenize(pattern):
    tokens = []
    i = 0
    while i < len(pattern):
        if pattern[i] in SPECIAL_SYMBOLS:
            tokens.append(pattern[i])
        elif pattern[i] in GROUP_SYMBOLS:
            j = i + 1
            while j < len(pattern) and pattern[j] != ']':
                j += 1
            tokens.append(pattern[i:j+1])
            i = j
        elif pattern[i] == ESCAPE_CHARACTER:
            tokens.append(pattern[i: i+2])
            i += 1
        else:
            tokens.append(pattern[i])
        i += 1
    return tokens


def test_token():
    assert tokenize('a[bc]d') == ['a', '[bc]', 'd']
    assert tokenize('a[bc]d[ef]') == ['a', '[bc]', 'd', '[ef]']
    assert tokenize("a*") == ['a', '*']



class Node:
    def __init__(self, value, left=None, right=None):
        self.value = value
        self.left = left
        self.right = right

def parse(tokens):
    stack = []
    for token in tokens:
        if token == '|':
            right = stack.pop()
            left = stack.pop()
            stack.append(Node('|', left, right))
        elif token == '*':
            left = stack.pop()
            stack.append(Node('*', left))
        elif token == '+':
            left = stack.pop()
            stack.append(Node('+', left))
        elif token == '?':
            left = stack.pop()
            stack.append(Node('?', left))
        else:
            stack.append(Node(token))


def run_tests():
    test_token()

if __name__ == '__main__':
    run_tests()
    print('All tests passed')