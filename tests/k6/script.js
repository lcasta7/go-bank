import http from 'k6/http';
import { sleep, check } from 'k6';
import { expect } from 'https://jslib.k6.io/k6chaijs/4.3.4.3/index.js';

export const options = {
	stages: [
		{
			target: 10,
			duration: '30s'
		},
	],
};



export default function() {
	const adminId = 1107
	const adminToken = logInUser(adminId)

	//test - get accounts
	const accountsUrl = 'http://host.docker.internal:3000/accounts';
	const accountsPayload = JSON.stringify({
		number: adminId
	});

	postRequest(accountsUrl, accountsPayload, adminToken)

	//test - transfer money
	//create new accounts
	const createAccountUrl = 'http://host.docker.internal:3000/account';
	const createFromAccountPayload = JSON.stringify({
		firstName: "k6Test",
		lastName: "FromAccount",
		password: "gobank",
		balance: getRandomInt(1000, 3000),
		admin_account: adminId
	});
	const createToAccountPayload = JSON.stringify({
		firstName: "k6Test",
		lastName: "ToAccount",
		password: "gobank",
		balance: getRandomInt(0, 20),
		admin_account: adminId
	});

	//call create and store numbers
	const fromAcc = postRequest(createAccountUrl, createFromAccountPayload, adminToken).json()
	const toAcc = postRequest(createAccountUrl, createToAccountPayload, adminToken).json()


	//login as fromAcc
	const fromAccToken = logInUser(fromAcc.number)

	//transfer
	const transferUrl = 'http://host.docker.internal:3000/transfer';
	const transferAmount = getRandomInt(20, 100);
	const transferPayload = JSON.stringify({
		from_number: fromAcc.number,
		to_number: toAcc.number,
		amount: transferAmount
	});

	//get updated accounts
	const fromAccUpdated = postRequest(transferUrl, transferPayload, fromAccToken).json();
	const toAccUpdated = getUpdatedToAcc(toAcc.number);

	deleteAccounts(fromAcc.id, toAcc.id, adminId, adminToken)

	expect(fromAcc.balance - transferAmount).to.equal(fromAccUpdated.balance)
	expect(toAcc.balance + transferAmount).to.equal(toAccUpdated.balance)
	sleep(5)
}
function getUpdatedToAcc(number) {
	const getAccUrl = 'http://host.docker.internal:3000/account/get';
	const getAccPayload = JSON.stringify({
		number: number,
	});
	const toAccToken = logInUser(number)

	return postRequest(getAccUrl, getAccPayload, toAccToken).json()
}


function logInUser(number) {
	const loginUrl = 'http://host.docker.internal:3000/login';
	const loginPayload = JSON.stringify({
		number: number,
		password: "gobank"
	});

	const loginParams = {
		headers: {
			'Content-Type': 'application/json',
		},
	};

	let loginRes = http.post(loginUrl, loginPayload, loginParams);
	check(loginRes, { "login status is 200": (r) => r.status === 200 });

	const token = loginRes.json('token');

	return token
}

function postRequest(endpoint, payload, token) {

	const params = {
		headers: {
			'Content-Type': 'application/json',
			'x-jwt-token': token
		}
	};

	const response = http.post(endpoint, payload, params);
	check(response, { [`${endpoint} status is 200`]: (r) => r.status === 200 });
	return response;
}

function deleteAccounts(fromAccId, toAccId, adminId, adminToken) {
	const deleteFromAccountUrl = `http://host.docker.internal:3000/account/${fromAccId}`;
	const deleteToAccountUrl = `http://host.docker.internal:3000/account/${toAccId}`;
	const deletePayload = JSON.stringify({
		admin_account: adminId

	});
	const deleteParams = {
		headers: {
			'Content-Type': 'application/json',
			'x-jwt-token': adminToken
		}
	};

	let response = http.del(deleteFromAccountUrl, deletePayload, deleteParams);
	check(response, { [`${deleteFromAccountUrl} status is 200`]: (r) => r.status === 200 });
	response = http.del(deleteToAccountUrl, deletePayload, deleteParams);
	check(response, { [`${deleteToAccountUrl} status is 200`]: (r) => r.status === 200 });
}

function getRandomInt(min, max) {
	// The maximum is exclusive and the minimum is inclusive
	min = Math.ceil(min);
	max = Math.floor(max);
	return Math.floor(Math.random() * (max - min) + min);
}
