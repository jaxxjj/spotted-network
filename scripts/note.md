export SIGNING_KEY_PATH="/Users/yimingchen/galxe/ready-on-repo/spotted-network/keys/signing/operator1.key.json"
export KEYSTORE_PASSWORD="testpassword"
export P2P_KEY_64="CAESQKW/y8x4MBT09AySrCDS1HXvsFEGoXLwqvWOQUifZ90TvdsBG0rSgcjJTH8qWwRYRysJaZ+7Z4egLxvShvBnQys="

/home/ec2-user/spotted-network/keys/signing/operator2.key.json
export P2P_KEY_64="CAESQHGMebvS8Wf6IZZh40yacCPzXhRlKqJCGfPySZyCFid6EdbnbwgelZkcZbllzWAZFfrdV/dcf2poB1OySA2mV0I="

go build -o spotted cmd/operator/main.go

export PATH=$PATH:~/bin

Generated P2P key for operator1:
  Private key (base64): CAESQKW/y8x4MBT09AySrCDS1HXvsFEGoXLwqvWOQUifZ90TvdsBG0rSgcjJTH8qWwRYRysJaZ+7Z4egLxvShvBnQys=
  Public key (base64): CAESIL3bARtK0oHIyUx/KlsEWEcrCWmfu2eHoC8b0obwZ0Mr
  P2PKey: 0x310c8425b620980dcfcf756e46572bb6ac80eb07
  PeerId: 12D3KooWNbUurxoy5Qn7hSRi5dvMdaeEFZQavacg253npoiuSJ9p

Generated P2P key for operator2:
  Private key (base64): CAESQHGMebvS8Wf6IZZh40yacCPzXhRlKqJCGfPySZyCFid6EdbnbwgelZkcZbllzWAZFfrdV/dcf2poB1OySA2mV0I=
  Public key (base64): CAESIBHW528IHpWZHGW5Zc1gGRX63Vf3XH9qaAdTskgNpldC
  P2PKey: 0x01078ffbf1de436d6f429f5ce6be8fd9d6e16165
  PeerId: 12D3KooWB21ALruLNKbH5vLjDjiG1mM9XVnCJMGdsdo2LKvoqneD

Generated P2P key for operator3:
  Private key (base64): CAESQM5ltPHuttHq7/HHHHymN5A/XSDKt5EPOwGWor2H3k0PXckF23DDwxzmdOhEtOy5f8szIAYWqSFH8cIlICumemo=
  Public key (base64): CAESIF3JBdtww8Mc5nToRLTsuX/LMyAGFqkhR/HCJSArpnpq
  P2PKey: 0x67aa23adde2459a1620be2ea28982310597521b0
  PeerId: 12D3KooWG8TsS8YsbnfArhnQznH6Rmumt6fwE9J66ZN8tF7n9GVf

```solidity
    address public constant OPERATOR_1 = address(0xCf593639B34CaE0ea3217dA27014ab5FbBAc8342);
    address public constant OPERATOR_2 = address(0xCCE3B4EC7681B4EcF5fD5b50e562A88a33E5137B);
    address public constant OPERATOR_3 = address(0xFE6B5379E861C79dB03eb3a01F3F1892FC4141D5);
    address public constant SIGNING_KEY_1 = address(0x9F0D8BAC11C5693a290527f09434b86651c66Bf2);
    address public constant SIGNING_KEY_2 = address(0xeBBAce05Db3D717A5BA82EAB8AdE712dFb151b13);
    address public constant SIGNING_KEY_3 = address(0x083739b681B85cc2c9e394471486321D6446b25b);
    address public constant P2P_KEY_1 = address(0x310C8425b620980DCFcf756e46572bb6ac80Eb07);
    address public constant P2P_KEY_2 = address(0x01078ffBf1De436d6f429f5Ce6Be8Fd9D6E16165);
    address public constant P2P_KEY_3 = address(0x67aa23adde2459a1620BE2Ea28982310597521b0);
```